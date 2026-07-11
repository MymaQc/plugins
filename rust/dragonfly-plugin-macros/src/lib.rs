use proc_macro::TokenStream;
use quote::quote;
use syn::{
    Attribute, Data, DeriveInput, Fields, FnArg, ImplItem, ItemImpl, LitStr, Type,
    parse_macro_input,
};

fn command_attribute(attributes: &[Attribute], key: &str) -> syn::Result<Option<String>> {
    let mut value = None;
    for attribute in attributes
        .iter()
        .filter(|attribute| attribute.path().is_ident("command"))
    {
        attribute.parse_nested_meta(|meta| {
            if meta.path.is_ident("name")
                || meta.path.is_ident("description")
                || meta.path.is_ident("value")
            {
                let parsed = meta.value()?.parse::<LitStr>()?.value();
                if meta.path.is_ident(key) {
                    value = Some(parsed);
                }
                Ok(())
            } else {
                Err(meta.error("unknown command option"))
            }
        })?;
    }
    Ok(value)
}

fn kebab_case(value: &str) -> String {
    let mut output = String::new();
    for (index, character) in value.chars().enumerate() {
        if character.is_uppercase() {
            if index != 0 {
                output.push('-');
            }
            output.extend(character.to_lowercase());
        } else if character == '_' {
            output.push('-');
        } else {
            output.push(character);
        }
    }
    output
}

fn type_name(value: &Type) -> Option<String> {
    let Type::Path(path) = value else {
        return None;
    };
    path.path
        .segments
        .last()
        .map(|segment| segment.ident.to_string())
}

fn parse_scalar(field: &syn::Ident, ty: &Type) -> proc_macro2::TokenStream {
    quote! {
        let raw = parts.next().ok_or_else(|| ::dragonfly_plugin::CommandParseError::new(
            concat!("missing command argument ", stringify!(#field))
        ))?;
        let #field = raw.parse::<#ty>().map_err(|_| {
            ::dragonfly_plugin::CommandParseError::new(format!("invalid {}: {raw}", stringify!(#field)))
        })?;
    }
}

#[proc_macro_derive(CommandEnum, attributes(command))]
pub fn command_enum(input: TokenStream) -> TokenStream {
    let input = parse_macro_input!(input as DeriveInput);
    match expand_command_enum(input) {
        Ok(output) => output.into(),
        Err(error) => error.into_compile_error().into(),
    }
}

fn expand_command_enum(input: DeriveInput) -> syn::Result<proc_macro2::TokenStream> {
    let name = input.ident;
    let Data::Enum(data) = input.data else {
        return Err(syn::Error::new_spanned(
            name,
            "CommandEnum requires an enum",
        ));
    };
    let mut values = Vec::new();
    let mut variants = Vec::new();
    for variant in data.variants {
        if !matches!(variant.fields, Fields::Unit) {
            return Err(syn::Error::new_spanned(
                variant,
                "command enum variants must be unit variants",
            ));
        }
        let value = command_attribute(&variant.attrs, "value")?
            .unwrap_or_else(|| kebab_case(&variant.ident.to_string()));
        values.push(value);
        variants.push(variant.ident);
    }
    Ok(quote! {
        impl ::dragonfly_plugin::CommandEnum for #name {
            const VALUES: &'static [::dragonfly_plugin::CommandValue] = &[
                #(::dragonfly_plugin::CommandValue::new(#values)),*
            ];

            fn parse(value: &str) -> ::core::option::Option<Self> {
                #(if value.eq_ignore_ascii_case(#values) {
                    return ::core::option::Option::Some(Self::#variants);
                })*
                ::core::option::Option::None
            }
        }
    })
}

#[proc_macro_derive(Command, attributes(command))]
pub fn command(input: TokenStream) -> TokenStream {
    let input = parse_macro_input!(input as DeriveInput);
    match expand_command(input) {
        Ok(output) => output.into(),
        Err(error) => error.into_compile_error().into(),
    }
}

fn expand_command(input: DeriveInput) -> syn::Result<proc_macro2::TokenStream> {
    let command_name = command_attribute(&input.attrs, "name")?
        .unwrap_or_else(|| kebab_case(&input.ident.to_string()));
    let description = command_attribute(&input.attrs, "description")?.unwrap_or_default();
    let name = input.ident;
    let Data::Enum(data) = input.data else {
        return Err(syn::Error::new_spanned(
            name,
            "Command requires an enum whose variants are subcommands",
        ));
    };
    let mut overloads = Vec::new();
    let mut parsers = Vec::new();
    for variant in data.variants {
        let variant_name = command_attribute(&variant.attrs, "name")?
            .unwrap_or_else(|| kebab_case(&variant.ident.to_string()));
        let variant_ident = variant.ident;
        let (parameters, field_names, reads): (Vec<_>, Vec<_>, Vec<_>) = match variant.fields {
            Fields::Unit => (Vec::new(), Vec::new(), Vec::new()),
            Fields::Named(fields) => {
                let mut parameters = Vec::new();
                let mut names = Vec::new();
                let mut reads = Vec::new();
                for field in fields.named {
                    let field_name = field.ident.expect("named field");
                    let parameter_name = command_attribute(&field.attrs, "name")?
                        .unwrap_or_else(|| kebab_case(&field_name.to_string()));
                    let field_type = field.ty;
                    let field_type_name = type_name(&field_type);
                    let (parameter, read) = match field_type_name.as_deref() {
                        Some("String") => (
                            quote!(::dragonfly_plugin::CommandParameter::string(#parameter_name)),
                            quote! {
                                let #field_name = parts.next().ok_or_else(|| {
                                    ::dragonfly_plugin::CommandParseError::new(
                                        concat!("missing command argument ", stringify!(#field_name))
                                    )
                                })?.to_owned();
                            },
                        ),
                        Some("bool") => (
                            quote!(::dragonfly_plugin::CommandParameter::boolean(#parameter_name)),
                            parse_scalar(&field_name, &field_type),
                        ),
                        Some("f32" | "f64") => (
                            quote!(::dragonfly_plugin::CommandParameter::float(#parameter_name)),
                            parse_scalar(&field_name, &field_type),
                        ),
                        Some(
                            "i8" | "i16" | "i32" | "i64" | "isize" | "u8" | "u16" | "u32" | "u64"
                            | "usize",
                        ) => (
                            quote!(::dragonfly_plugin::CommandParameter::integer(#parameter_name)),
                            parse_scalar(&field_name, &field_type),
                        ),
                        _ => (
                            quote! {
                                ::dragonfly_plugin::CommandParameter::enumeration(
                                    #parameter_name,
                                    <#field_type as ::dragonfly_plugin::CommandEnum>::VALUES,
                                )
                            },
                            quote! {
                                let raw = parts.next().ok_or_else(|| ::dragonfly_plugin::CommandParseError::new(
                                    concat!("missing command argument ", stringify!(#field_name))
                                ))?;
                                let #field_name = <#field_type as ::dragonfly_plugin::CommandEnum>::parse(raw).ok_or_else(|| {
                                    ::dragonfly_plugin::CommandParseError::new(format!("invalid {}: {raw}", stringify!(#field_name)))
                                })?;
                            },
                        ),
                    };
                    parameters.push(parameter);
                    names.push(field_name);
                    reads.push(read);
                }
                (parameters, names, reads)
            }
            Fields::Unnamed(fields) => {
                return Err(syn::Error::new_spanned(
                    fields,
                    "command variants must use named fields",
                ));
            }
        };
        if parameters.len() > 3 {
            return Err(syn::Error::new_spanned(
                &variant_ident,
                "a command supports at most three enum fields after its subcommand",
            ));
        }
        overloads.push(quote! {
            ::dragonfly_plugin::CommandOverload::new(&[
                ::dragonfly_plugin::CommandParameter::subcommand(#variant_name),
                #(#parameters),*
            ])
        });
        let construct = if field_names.is_empty() {
            quote!(Self::#variant_ident)
        } else {
            quote!(Self::#variant_ident { #(#field_names),* })
        };
        parsers.push(quote! {
            if subcommand.eq_ignore_ascii_case(#variant_name) {
                #(#reads)*
                if let Some(extra) = parts.next() {
                    return Err(::dragonfly_plugin::CommandParseError::new(format!("unexpected command argument: {extra}")));
                }
                return Ok(#construct);
            }
        });
    }
    Ok(quote! {
        impl ::dragonfly_plugin::CommandDefinition for #name {
            const COMMAND: ::dragonfly_plugin::Command =
                ::dragonfly_plugin::Command::new(#command_name, #description).with_overloads(&[
                    #(#overloads),*
                ]);

            fn parse(arguments: &str) -> ::core::result::Result<Self, ::dragonfly_plugin::CommandParseError> {
                let mut parts = arguments.split_whitespace();
                let subcommand = parts.next().ok_or_else(|| {
                    ::dragonfly_plugin::CommandParseError::new("missing subcommand")
                })?;
                #(#parsers)*
                Err(::dragonfly_plugin::CommandParseError::new(format!("unknown subcommand: {subcommand}")))
            }
        }
    })
}

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
    let mut implementation = parse_macro_input!(input as ItemImpl);
    let plugin_type = implementation.self_ty.clone();
    let mut command_methods = Vec::new();
    let mut retained = Vec::with_capacity(implementation.items.len());
    for item in implementation.items {
        match item {
            ImplItem::Fn(mut function)
                if function
                    .attrs
                    .iter()
                    .any(|attribute| attribute.path().is_ident("command")) =>
            {
                function
                    .attrs
                    .retain(|attribute| !attribute.path().is_ident("command"));
                command_methods.push(function);
            }
            item => retained.push(item),
        }
    }
    implementation.items = retained;
    if !command_methods.is_empty()
        && implementation.items.iter().any(|item| {
            matches!(item, ImplItem::Fn(function) if function.sig.ident == "commands" || function.sig.ident == "on_command")
        })
    {
        return syn::Error::new_spanned(
            &implementation,
            "#[command] methods cannot be mixed with manual commands() or on_command() handlers",
        )
        .into_compile_error()
        .into();
    }
    let mut command_types: Vec<Type> = Vec::new();
    let mut command_names = Vec::new();
    for function in &command_methods {
        let Some(FnArg::Typed(arguments)) = function.sig.inputs.iter().nth(2) else {
            return syn::Error::new_spanned(
                &function.sig,
                "a #[command] method must take &self, &mut CommandEvent, and a derived command value",
            )
            .into_compile_error()
            .into();
        };
        if function.sig.inputs.len() != 3 {
            return syn::Error::new_spanned(
                &function.sig,
                "a #[command] method must take exactly three arguments",
            )
            .into_compile_error()
            .into();
        }
        command_types.push((*arguments.ty).clone());
        command_names.push(function.sig.ident.clone());
    }
    if !command_methods.is_empty() {
        let dispatch_arms = command_types.iter().zip(command_names.iter()).enumerate().map(
            |(index, (command_type, method))| quote! {
                #index => match <#command_type as ::dragonfly_plugin::CommandDefinition>::parse(event.arguments()) {
                    Ok(command) => self.#method(event, command),
                    Err(error) => {
                        let _ = event.fail(&error.to_string());
                    }
                },
            },
        );
        implementation.items.push(syn::parse_quote! {
            fn commands(&self) -> &'static [::dragonfly_plugin::Command] {
                const COMMANDS: &[::dragonfly_plugin::Command] = &[
                    #(<#command_types as ::dragonfly_plugin::CommandDefinition>::COMMAND),*
                ];
                COMMANDS
            }
        });
        implementation.items.push(syn::parse_quote! {
            fn on_command(&self, command: usize, event: &mut ::dragonfly_plugin::CommandEvent<'_>) {
                match command {
                    #(#dispatch_arms)*
                    _ => {}
                }
            }
        });
    }
    let inherent_commands = if command_methods.is_empty() {
        quote!()
    } else {
        quote! {
            impl #plugin_type {
                #(#command_methods)*
            }
        }
    };
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
        #inherent_commands

        #[doc(hidden)]
        mod __dragonfly_plugin_export {
            use super::*;

            type PluginType = #plugin_type;
            const PLUGIN_ID: &[u8] = env!("CARGO_PKG_NAME").as_bytes();

            unsafe extern "C" fn create() -> *mut ::dragonfly_plugin::__private::c_void {
                match ::std::panic::catch_unwind(|| <PluginType as ::core::default::Default>::default()) {
                    Ok(plugin) => ::std::boxed::Box::into_raw(::std::boxed::Box::new(plugin)).cast(),
                    Err(_) => ::core::ptr::null_mut(),
                }
            }

            unsafe extern "C" fn enable(instance: *mut ::dragonfly_plugin::__private::c_void) -> ::dragonfly_plugin::__private::sys::DfStatus {
                let plugin = unsafe { &*instance.cast::<PluginType>() };
                match ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                    <PluginType as ::dragonfly_plugin::Plugin>::on_enable(plugin);
                })) {
                    Ok(()) => ::dragonfly_plugin::__private::sys::DF_STATUS_OK,
                    Err(_) => ::dragonfly_plugin::__private::sys::DF_STATUS_ERROR,
                }
            }

            unsafe extern "C" fn disable(instance: *mut ::dragonfly_plugin::__private::c_void) -> ::dragonfly_plugin::__private::sys::DfStatus {
                let plugin = unsafe { &*instance.cast::<PluginType>() };
                match ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                    <PluginType as ::dragonfly_plugin::Plugin>::on_disable(plugin);
                })) {
                    Ok(()) => ::dragonfly_plugin::__private::sys::DF_STATUS_OK,
                    Err(_) => ::dragonfly_plugin::__private::sys::DF_STATUS_ERROR,
                }
            }

            unsafe extern "C" fn destroy(instance: *mut ::dragonfly_plugin::__private::c_void) {
                if !instance.is_null() {
                    let _ = ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                        drop(unsafe { ::std::boxed::Box::from_raw(instance.cast::<PluginType>()) });
                    }));
                }
            }

            unsafe extern "C" fn commands(
                instance: *mut ::dragonfly_plugin::__private::c_void,
                count: *mut u64,
            ) -> *const ::dragonfly_plugin::__private::sys::DfCommandDescriptor {
                if instance.is_null() || count.is_null() {
                    return ::core::ptr::null();
                }
                let plugin = unsafe { &*instance.cast::<PluginType>() };
                match ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                    <PluginType as ::dragonfly_plugin::Plugin>::commands(plugin)
                })) {
                    Ok(commands) => {
                        unsafe { *count = commands.len() as u64 };
                        commands.as_ptr().cast()
                    }
                    Err(_) => {
                        unsafe { *count = u64::MAX };
                        ::core::ptr::null()
                    }
                }
            }

            unsafe extern "C" fn handle_command(
                instance: *mut ::dragonfly_plugin::__private::c_void,
                command: u64,
                input: *const ::dragonfly_plugin::__private::sys::DfCommandInput,
                state: *mut ::dragonfly_plugin::__private::sys::DfCommandState,
            ) -> ::dragonfly_plugin::__private::sys::DfStatus {
                use ::dragonfly_plugin::__private::sys;
                if instance.is_null() || input.is_null() || state.is_null() {
                    return sys::DF_STATUS_ERROR;
                }
                let plugin = unsafe { &*instance.cast::<PluginType>() };
                if command as usize >= <PluginType as ::dragonfly_plugin::Plugin>::commands(plugin).len() {
                    return sys::DF_STATUS_ERROR;
                }
                let result = ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                    let mut event = unsafe { ::dragonfly_plugin::CommandEvent::from_raw(&*input, &mut *state) };
                    <PluginType as ::dragonfly_plugin::Plugin>::on_command(plugin, command as usize, &mut event);
                }));
                if result.is_ok() { sys::DF_STATUS_OK } else { sys::DF_STATUS_ERROR }
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
                    enable: Some(enable),
                    disable: Some(disable),
                    commands: Some(commands),
                    handle_command: Some(handle_command),
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
