use proc_macro::TokenStream;
use quote::quote;
use syn::{ItemImpl, parse_macro_input};

#[proc_macro_attribute]
pub fn plugin(attributes: TokenStream, input: TokenStream) -> TokenStream {
    if !attributes.is_empty() {
        return syn::Error::new(
            proc_macro2::Span::call_site(),
            "`#[plugin]` takes no arguments; identity comes from Cargo package metadata",
        )
        .into_compile_error()
        .into();
    }
    let implementation = parse_macro_input!(input as ItemImpl);
    let plugin_type = &implementation.self_ty;
    let handles_move = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_move"));
    let handles_chat = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_chat"));
    let subscriptions = u64::from(handles_move) | (u64::from(handles_chat) << 1);

    quote! {
        #implementation

        #[doc(hidden)]
        mod __dragonfly_plugin_export {
            use super::*;

            type PluginType = #plugin_type;
            const PLUGIN_ID: &[u8] = env!("CARGO_PKG_NAME").as_bytes();

            unsafe extern "C" fn create() -> *mut ::dragonfly_plugin::__private::c_void {
                match ::std::panic::catch_unwind(|| <PluginType as ::core::default::Default>::default()) {
                    Ok(plugin) => {
                        if ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                            <PluginType as ::dragonfly_plugin::Plugin>::on_enable(&plugin);
                        })).is_err() {
                            return ::core::ptr::null_mut();
                        }
                        ::std::boxed::Box::into_raw(::std::boxed::Box::new(plugin)).cast()
                    }
                    Err(_) => ::core::ptr::null_mut(),
                }
            }

            unsafe extern "C" fn destroy(instance: *mut ::dragonfly_plugin::__private::c_void) {
                if !instance.is_null() {
                    let _ = ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                        let plugin = unsafe { ::std::boxed::Box::from_raw(instance.cast::<PluginType>()) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_disable(&plugin);
                        drop(plugin);
                    }));
                }
            }

            unsafe extern "C" fn handle_event(
                instance: *mut ::dragonfly_plugin::__private::c_void,
                event_id: ::dragonfly_plugin::__private::sys::DfEventId,
                input: *const ::dragonfly_plugin::__private::c_void,
                state: *mut ::dragonfly_plugin::__private::c_void,
            ) -> ::dragonfly_plugin::__private::sys::DfStatus {
                use ::dragonfly_plugin::__private::sys;
                if instance.is_null() || input.is_null() || state.is_null() {
                    return sys::DF_STATUS_ERROR;
                }
                let result = ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| match event_id {
                    sys::DF_EVENT_PLAYER_MOVE => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerMoveInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerMoveState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerMoveEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_move(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_CHAT => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerChatInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerChatState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerChatEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_chat(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    _ => sys::DF_STATUS_ERROR,
                }));
                result.unwrap_or(sys::DF_STATUS_ERROR)
            }

            static API: ::dragonfly_plugin::__private::sys::DfPluginApiV1 =
                ::dragonfly_plugin::__private::sys::DfPluginApiV1 {
                    header: ::dragonfly_plugin::__private::sys::DfAbiHeader {
                        abi_version: ::dragonfly_plugin::__private::sys::DF_ABI_VERSION,
                        struct_size: ::core::mem::size_of::<::dragonfly_plugin::__private::sys::DfPluginApiV1>() as u32,
                        subscriptions: #subscriptions,
                    },
                    plugin_id: ::dragonfly_plugin::__private::sys::DfStringView {
                        data: PLUGIN_ID.as_ptr(),
                        len: PLUGIN_ID.len() as u64,
                    },
                    create: Some(create),
                    destroy: Some(destroy),
                    handle_event: Some(handle_event),
                };

            #[unsafe(no_mangle)]
            pub extern "C" fn df_plugin_entry_v1() -> *const ::dragonfly_plugin::__private::sys::DfPluginApiV1 {
                &API
            }
        }
    }
    .into()
}
