use proc_macro::TokenStream;
use quote::quote;
use std::collections::BTreeMap;
use syn::{
    Attribute, Data, DeriveInput, Fields, FnArg, GenericArgument, ImplItem, ItemImpl, LitStr, Meta,
    Pat, PathArguments, Type, parse_macro_input,
};

fn command_method_path(attributes: &[Attribute]) -> syn::Result<Option<String>> {
    let Some(attribute) = attributes
        .iter()
        .find(|attribute| attribute.path().is_ident("command"))
    else {
        return Ok(None);
    };
    if matches!(attribute.meta, Meta::Path(_)) {
        return Ok(None);
    }
    Ok(Some(attribute.parse_args::<LitStr>()?.value()))
}

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

fn generic_argument(value: &Type, expected: &str) -> Option<Type> {
    let Type::Path(path) = value else {
        return None;
    };
    let segment = path.path.segments.last()?;
    if segment.ident != expected {
        return None;
    }
    let PathArguments::AngleBracketed(arguments) = &segment.arguments else {
        return None;
    };
    match arguments.args.first()? {
        GenericArgument::Type(provider) => Some(provider.clone()),
        _ => None,
    }
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
    let mut dynamic_options = Vec::new();
    for (overload_index, variant) in data.variants.into_iter().enumerate() {
        let variant_name = command_attribute(&variant.attrs, "name")?
            .unwrap_or_else(|| kebab_case(&variant.ident.to_string()));
        let variant_ident = variant.ident;
        let (parameters, field_names, reads, named_variant): (Vec<_>, Vec<_>, Vec<_>, bool) =
            match variant.fields {
                Fields::Unit => (Vec::new(), Vec::new(), Vec::new(), false),
                Fields::Named(fields) => {
                    let mut parameters = Vec::new();
                    let mut names = Vec::new();
                    let mut reads = Vec::new();
                    let mut seen_optional = false;
                    let field_count = fields.named.len();
                    for (field_index, field) in fields.named.into_iter().enumerate() {
                        let field_name = field.ident.expect("named field");
                        let parameter_name = command_attribute(&field.attrs, "name")?
                            .unwrap_or_else(|| kebab_case(&field_name.to_string()));
                        let declared_type = field.ty;
                        let optional_type = generic_argument(&declared_type, "Option");
                        let optional = optional_type.is_some();
                        if seen_optional && !optional {
                            return Err(syn::Error::new_spanned(
                                &field_name,
                                "required command fields cannot follow Option fields",
                            ));
                        }
                        seen_optional |= optional;
                        let field_type = optional_type.unwrap_or_else(|| declared_type.clone());
                        let provider = generic_argument(&field_type, "Dynamic");
                        let field_type_name = type_name(&field_type);
                        let varargs = field_type_name.as_deref() == Some("Varargs");
                        let (mut parameter, mut read) = if varargs {
                            if field_index + 1 != field_count {
                                return Err(syn::Error::new_spanned(
                                    &field_name,
                                    "Varargs must be the final command field",
                                ));
                            }
                            (
                                quote!(::dragonfly_plugin::CommandParameter::raw_text(#parameter_name)),
                                quote! {
                                let #field_name = ::dragonfly_plugin::Varargs::new(
                                    parts.by_ref().collect::<Vec<_>>().join(" ")
                                );
                                },
                            )
                        } else if let Some(provider) = provider {
                            let parameter_index = field_index + 1;
                            dynamic_options.push(quote! {
                            (#overload_index, #parameter_index) => {
                                return Some(<#provider as ::dragonfly_plugin::DynamicCommandEnum>::options(source));
                            }
                        });
                            (
                                quote!(::dragonfly_plugin::CommandParameter::dynamic_enum(#parameter_name)),
                                quote! {
                                    let #field_name = ::dragonfly_plugin::Dynamic::new(
                                        parts.next().ok_or_else(|| {
                                            ::dragonfly_plugin::CommandParseError::new(
                                                concat!("missing command argument ", stringify!(#field_name))
                                            )
                                        })?
                                    );
                                },
                            )
                        } else {
                            match field_type_name.as_deref() {
                                Some("Player") => (
                                    quote!(::dragonfly_plugin::CommandParameter::player(#parameter_name)),
                                    quote! {
                                        let raw = parts.next().ok_or_else(|| {
                                            ::dragonfly_plugin::CommandParseError::new(
                                                concat!("missing command argument ", stringify!(#field_name))
                                            )
                                        })?;
                                        let #field_name = ::dragonfly_plugin::Player::from_command_argument(raw)?;
                                    },
                                ),
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
                                    "i8" | "i16" | "i32" | "i64" | "isize" | "u8" | "u16" | "u32"
                                    | "u64" | "usize",
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
                            }
                        };
                        if optional {
                            parameter = quote!((#parameter).optional());
                            read = quote! {
                                let #field_name = if parts.clone().next().is_none() {
                                    None
                                } else {
                                    Some({
                                        #read
                                        #field_name
                                    })
                                };
                            };
                        }
                        parameters.push(parameter);
                        names.push(field_name);
                        reads.push(read);
                    }
                    (parameters, names, reads, true)
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
        let construct = if named_variant {
            quote!(Self::#variant_ident { #(#field_names),* })
        } else {
            quote!(Self::#variant_ident)
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

            fn dynamic_options(
                overload: usize,
                parameter: usize,
                source: ::dragonfly_plugin::CommandSource<'_>,
            ) -> Option<Vec<String>> {
                match (overload, parameter) {
                    #(#dynamic_options)*
                    _ => None,
                }
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
    let mut legacy_methods = Vec::new();
    let mut direct_groups: BTreeMap<String, Vec<(String, syn::ImplItemFn)>> = BTreeMap::new();
    let mut retained = Vec::with_capacity(implementation.items.len());
    for item in implementation.items {
        match item {
            ImplItem::Fn(mut function)
                if function
                    .attrs
                    .iter()
                    .any(|attribute| attribute.path().is_ident("command")) =>
            {
                let path = match command_method_path(&function.attrs) {
                    Ok(path) => path,
                    Err(error) => return error.into_compile_error().into(),
                };
                function
                    .attrs
                    .retain(|attribute| !attribute.path().is_ident("command"));
                if let Some(path) = path {
                    let parts: Vec<_> = path.split_whitespace().collect();
                    if parts.len() != 2 {
                        return syn::Error::new_spanned(
                            &function.sig,
                            "#[command(\"root subcommand\")] requires exactly two path components",
                        )
                        .into_compile_error()
                        .into();
                    }
                    direct_groups
                        .entry(parts[0].to_owned())
                        .or_default()
                        .push((parts[1].to_owned(), function));
                } else {
                    legacy_methods.push(function);
                }
            }
            item => retained.push(item),
        }
    }
    implementation.items = retained;
    let has_commands = !legacy_methods.is_empty() || !direct_groups.is_empty();
    if has_commands
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
    let mut command_types: Vec<proc_macro2::TokenStream> = Vec::new();
    let mut dispatch_arms = Vec::new();
    let mut generated_commands = Vec::new();
    let mut inherent_methods = Vec::new();
    for function in legacy_methods {
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
        let command_type = (*arguments.ty).clone();
        let method = function.sig.ident.clone();
        let index = command_types.len();
        command_types.push(quote!(#command_type));
        dispatch_arms.push(quote! {
            #index => match <#command_type as ::dragonfly_plugin::CommandDefinition>::parse(event.arguments()) {
                Ok(command) => self.#method(event, command),
                Err(error) => {
                    let _ = event.fail(&error.to_string());
                }
            },
        });
        inherent_methods.push(function);
    }
    for (group_index, (root, methods)) in direct_groups.into_iter().enumerate() {
        let command_type = syn::Ident::new(
            &format!("__DragonflyDirectCommand{group_index}"),
            proc_macro2::Span::call_site(),
        );
        let description = format!("Runs {root} commands");
        let mut variants = Vec::new();
        let mut variant_dispatch = Vec::new();
        for (variant_index, (subcommand, function)) in methods.into_iter().enumerate() {
            if function.sig.inputs.len() < 2 {
                return syn::Error::new_spanned(
                    &function.sig,
                    "a direct #[command] method must take &self and &mut CommandEvent",
                )
                .into_compile_error()
                .into();
            }
            let variant = syn::Ident::new(
                &format!("Variant{variant_index}"),
                proc_macro2::Span::call_site(),
            );
            let method = function.sig.ident.clone();
            let mut fields = Vec::new();
            let mut field_names = Vec::new();
            for argument in function.sig.inputs.iter().skip(2) {
                let FnArg::Typed(argument) = argument else {
                    return syn::Error::new_spanned(
                        argument,
                        "command arguments must be named values",
                    )
                    .into_compile_error()
                    .into();
                };
                let Pat::Ident(pattern) = argument.pat.as_ref() else {
                    return syn::Error::new_spanned(
                        &argument.pat,
                        "command arguments must use identifier patterns",
                    )
                    .into_compile_error()
                    .into();
                };
                let field = pattern.ident.clone();
                let ty = argument.ty.clone();
                fields.push(quote!(#field: #ty));
                field_names.push(field);
            }
            variants.push(quote! {
                #[command(name = #subcommand)]
                #variant { #(#fields),* }
            });
            variant_dispatch.push(quote! {
                #command_type::#variant { #(#field_names),* } => self.#method(event, #(#field_names),*),
            });
            inherent_methods.push(function);
        }
        generated_commands.push(quote! {
            #[derive(::dragonfly_plugin::Command)]
            #[command(name = #root, description = #description)]
            enum #command_type {
                #(#variants),*
            }
        });
        let index = command_types.len();
        command_types.push(quote!(#command_type));
        dispatch_arms.push(quote! {
            #index => match <#command_type as ::dragonfly_plugin::CommandDefinition>::parse(event.arguments()) {
                Ok(command) => match command {
                    #(#variant_dispatch)*
                },
                Err(error) => {
                    let _ = event.fail(&error.to_string());
                }
            },
        });
    }
    if has_commands {
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
    let inherent_commands = if !has_commands {
        quote!()
    } else {
        quote! {
            impl #plugin_type {
                #(#inherent_methods)*
            }
        }
    };
    let dynamic_dispatch_arms: Vec<_> = command_types
        .iter()
        .enumerate()
        .map(|(index, command_type)| {
            quote! {
                #index => <#command_type as ::dragonfly_plugin::CommandDefinition>::dynamic_options(
                    overload as usize,
                    parameter as usize,
                    source,
                ),
            }
        })
        .collect();
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
        #(#generated_commands)*
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

            unsafe extern "C" fn command_enum_options(
                instance: *mut ::dragonfly_plugin::__private::c_void,
                command: u64,
                overload: u64,
                parameter: u64,
                context: *const ::dragonfly_plugin::__private::sys::DfCommandEnumContext,
                output: *mut ::dragonfly_plugin::__private::sys::DfStringBuffer,
            ) -> ::dragonfly_plugin::__private::sys::DfStatus {
                use ::dragonfly_plugin::__private::sys;
                if instance.is_null() || context.is_null() || output.is_null() {
                    return sys::DF_STATUS_ERROR;
                }
                let result = ::std::panic::catch_unwind(::std::panic::AssertUnwindSafe(|| {
                    let context = unsafe { &*context };
                    if context.source.len != 0 && context.source.data.is_null() {
                        return Err(());
                    }
                    let bytes = if context.source.len == 0 {
                        &[][..]
                    } else {
                        unsafe { ::core::slice::from_raw_parts(context.source.data, context.source.len as usize) }
                    };
                    let name = ::core::str::from_utf8(bytes).map_err(|_| ())?;
                    let online_players = if context.online_player_count == 0 {
                        &[][..]
                    } else {
                        if context.online_players.is_null() {
                            return Err(());
                        }
                        unsafe {
                            ::core::slice::from_raw_parts(
                                context.online_players,
                                context.online_player_count as usize,
                            )
                        }
                    };
                    let source = ::dragonfly_plugin::CommandSource::new(name, online_players);
                    let options = match command as usize {
                        #(#dynamic_dispatch_arms)*
                        _ => None,
                    }.ok_or(())?;
                    unsafe { (*output).len = 0 };
                    ::dragonfly_plugin::write_dynamic_options(options, unsafe { &mut *output }).map_err(|_| ())
                }));
                match result {
                    Ok(Ok(())) => sys::DF_STATUS_OK,
                    _ => sys::DF_STATUS_ERROR,
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
                    enable: Some(enable),
                    disable: Some(disable),
                    commands: Some(commands),
                    handle_command: Some(handle_command),
                    command_enum_options: Some(command_enum_options),
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
