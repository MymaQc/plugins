use proc_macro::TokenStream;
use quote::quote;
use syn::{Expr, ItemImpl, Lit, MetaNameValue, Token, parse::Parser, parse_macro_input};

#[proc_macro_attribute]
pub fn plugin(attributes: TokenStream, input: TokenStream) -> TokenStream {
    let parser = syn::punctuated::Punctuated::<MetaNameValue, Token![,]>::parse_terminated;
    let attributes = match parser.parse(attributes) {
        Ok(attributes) => attributes,
        Err(error) => return error.into_compile_error().into(),
    };
    let mut plugin_id = None;
    for attribute in attributes {
        if !attribute.path.is_ident("id") {
            return syn::Error::new_spanned(attribute.path, "expected `id = \"namespace:name\"`")
                .into_compile_error()
                .into();
        }
        let Expr::Lit(expression) = attribute.value else {
            return syn::Error::new_spanned(attribute.value, "plugin ID must be a string literal")
                .into_compile_error()
                .into();
        };
        let Lit::Str(value) = expression.lit else {
            return syn::Error::new_spanned(expression, "plugin ID must be a string literal")
                .into_compile_error()
                .into();
        };
        plugin_id = Some(value);
    }
    let Some(plugin_id) = plugin_id else {
        return syn::Error::new(
            proc_macro2::Span::call_site(),
            "missing `id = \"namespace:name\"`",
        )
        .into_compile_error()
        .into();
    };
    let implementation = parse_macro_input!(input as ItemImpl);
    let plugin_type = &implementation.self_ty;
    let handles_move = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_move"));
    let subscriptions = if handles_move { 1u64 } else { 0 };

    quote! {
        #implementation

        #[doc(hidden)]
        mod __dragonfly_plugin_export {
            use super::*;

            type PluginType = #plugin_type;
            const PLUGIN_ID: &[u8] = #plugin_id.as_bytes();

            unsafe extern "C" fn create() -> *mut ::dragonfly_plugin::__private::c_void {
                match ::std::panic::catch_unwind(|| <PluginType as ::core::default::Default>::default()) {
                    Ok(plugin) => ::std::boxed::Box::into_raw(::std::boxed::Box::new(plugin)).cast(),
                    Err(_) => ::core::ptr::null_mut(),
                }
            }

            unsafe extern "C" fn destroy(instance: *mut ::dragonfly_plugin::__private::c_void) {
                if !instance.is_null() {
                    let _ = ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                        drop(unsafe { ::std::boxed::Box::from_raw(instance.cast::<PluginType>()) });
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
