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

fn subcommand_name(attributes: &[Attribute]) -> syn::Result<Option<String>> {
    let Some(attribute) = attributes
        .iter()
        .find(|attribute| attribute.path().is_ident("subcommand"))
    else {
        return Ok(None);
    };
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
            } else if meta.path.is_ident("root") {
                Ok(())
            } else {
                Err(meta.error("unknown command option"))
            }
        })?;
    }
    Ok(value)
}

fn command_flag(attributes: &[Attribute], key: &str) -> syn::Result<bool> {
    let mut found = false;
    for attribute in attributes
        .iter()
        .filter(|attribute| attribute.path().is_ident("command"))
    {
        attribute.parse_nested_meta(|meta| {
            if meta.path.is_ident("root") {
                found |= key == "root";
                return Ok(());
            }
            if meta.path.is_ident("name")
                || meta.path.is_ident("description")
                || meta.path.is_ident("value")
            {
                let _ = meta.value()?.parse::<LitStr>()?;
                return Ok(());
            }
            Err(meta.error("unknown command option"))
        })?;
    }
    Ok(found)
}

#[derive(Clone, Copy)]
enum ContextRestriction {
    Any,
    Player,
    Console,
}

fn context_restriction(function: &syn::ImplItemFn) -> ContextRestriction {
    let Some(FnArg::Typed(argument)) = function.sig.inputs.iter().nth(1) else {
        return ContextRestriction::Any;
    };
    let Type::Reference(reference) = argument.ty.as_ref() else {
        return ContextRestriction::Any;
    };
    let Type::Path(path) = reference.elem.as_ref() else {
        return ContextRestriction::Any;
    };
    let Some(segment) = path.path.segments.last() else {
        return ContextRestriction::Any;
    };
    let PathArguments::AngleBracketed(arguments) = &segment.arguments else {
        return ContextRestriction::Any;
    };
    for argument in &arguments.args {
        let GenericArgument::Type(Type::Path(path)) = argument else {
            continue;
        };
        if path.path.is_ident("Player") {
            return ContextRestriction::Player;
        }
        if path.path.is_ident("Console") {
            return ContextRestriction::Console;
        }
    }
    ContextRestriction::Any
}

fn restricted_call(
    restriction: ContextRestriction,
    call: proc_macro2::TokenStream,
) -> proc_macro2::TokenStream {
    match restriction {
        ContextRestriction::Any => call,
        ContextRestriction::Player => quote! {
            if let Some(mut restricted_context) = context.player_context() {
                #call
            } else {
                context.fail("This command can only be used by a player.");
            }
        },
        ContextRestriction::Console => quote! {
            if let Some(mut restricted_context) = context.console_context() {
                #call
            } else {
                context.fail("This command can only be used by the console.");
            }
        },
    }
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
    let mut root_parser = None;
    let mut dynamic_options = Vec::new();
    for (overload_index, variant) in data.variants.into_iter().enumerate() {
        let variant_name = command_attribute(&variant.attrs, "name")?
            .unwrap_or_else(|| kebab_case(&variant.ident.to_string()));
        let is_root = command_flag(&variant.attrs, "root")?;
        if is_root && root_parser.is_some() {
            return Err(syn::Error::new_spanned(
                &variant.ident,
                "a command may only have one root runnable",
            ));
        }
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
                            let parameter_index = field_index + usize::from(!is_root);
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
        let maximum_parameters = if is_root { 4 } else { 3 };
        if parameters.len() > maximum_parameters {
            return Err(syn::Error::new_spanned(
                &variant_ident,
                "a command supports at most three enum fields after its subcommand",
            ));
        }
        if is_root {
            overloads.push(quote! {
                ::dragonfly_plugin::CommandOverload::new(&[#(#parameters),*])
            });
        } else {
            overloads.push(quote! {
                ::dragonfly_plugin::CommandOverload::new(&[
                    ::dragonfly_plugin::CommandParameter::subcommand(#variant_name),
                    #(#parameters),*
                ])
            });
        }
        let construct = if named_variant {
            quote!(Self::#variant_ident { #(#field_names),* })
        } else {
            quote!(Self::#variant_ident)
        };
        let parser = quote! {
                #(#reads)*
                if let Some(extra) = parts.next() {
                    return Err(::dragonfly_plugin::CommandParseError::new(format!("unexpected command argument: {extra}")));
                }
                return Ok(#construct);
        };
        if is_root {
            root_parser = Some(parser);
        } else {
            parsers.push(quote! {
            if subcommand.is_some_and(|value| value.eq_ignore_ascii_case(#variant_name)) {
                let _ = parts.next();
                #parser
            }
            });
        }
    }
    let root_parser = root_parser.unwrap_or_else(|| quote! {
        let subcommand = subcommand.unwrap_or("<missing>");
        return Err(::dragonfly_plugin::CommandParseError::new(format!("unknown subcommand: {subcommand}")));
    });
    Ok(quote! {
        impl ::dragonfly_plugin::CommandDefinition for #name {
            const COMMAND: ::dragonfly_plugin::Command =
                ::dragonfly_plugin::Command::new(#command_name, #description).with_overloads(&[
                    #(#overloads),*
                ]);

            fn parse(arguments: &str) -> ::core::result::Result<Self, ::dragonfly_plugin::CommandParseError> {
                let mut parts = arguments.split_whitespace();
                let subcommand = parts.clone().next();
                #(#parsers)*
                #root_parser
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
    let command_root = match command_method_path(&implementation.attrs) {
        Ok(root) => root,
        Err(error) => return error.into_compile_error().into(),
    };
    implementation
        .attrs
        .retain(|attribute| !attribute.path().is_ident("command"));
    let plugin_type = implementation.self_ty.clone();
    let mut legacy_methods = Vec::new();
    let mut direct_groups: BTreeMap<String, Vec<(String, syn::ImplItemFn)>> = BTreeMap::new();
    let mut retained = Vec::with_capacity(implementation.items.len());
    for item in implementation.items {
        match item {
            ImplItem::Fn(mut function)
                if function.attrs.iter().any(|attribute| {
                    attribute.path().is_ident("command") || attribute.path().is_ident("subcommand")
                }) =>
            {
                let subcommand = match subcommand_name(&function.attrs) {
                    Ok(subcommand) => subcommand,
                    Err(error) => return error.into_compile_error().into(),
                };
                let path = match command_method_path(&function.attrs) {
                    Ok(path) => path,
                    Err(error) => return error.into_compile_error().into(),
                };
                function.attrs.retain(|attribute| {
                    !attribute.path().is_ident("command")
                        && !attribute.path().is_ident("subcommand")
                });
                if let Some(subcommand) = subcommand {
                    let Some(root) = &command_root else {
                        return syn::Error::new_spanned(
                            &function.sig,
                            "#[subcommand] requires #[command(\"root\")] on the plugin impl",
                        )
                        .into_compile_error()
                        .into();
                    };
                    if path.is_some() {
                        return syn::Error::new_spanned(
                            &function.sig,
                            "a method cannot be both #[command] and #[subcommand]",
                        )
                        .into_compile_error()
                        .into();
                    }
                    direct_groups
                        .entry(root.clone())
                        .or_default()
                        .push((subcommand, function));
                } else if let Some(root) = &command_root {
                    let subcommand = path.unwrap_or_default();
                    if subcommand.split_whitespace().count() > 1 {
                        return syn::Error::new_spanned(
                            &function.sig,
                            "method command attributes under a root contain only the subcommand",
                        )
                        .into_compile_error()
                        .into();
                    }
                    direct_groups
                        .entry(root.clone())
                        .or_default()
                        .push((subcommand, function));
                } else if let Some(path) = path {
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
                "a #[command] method must take &self, &mut Context, and a derived command value",
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
        let restriction = context_restriction(&function);
        let call = match restriction {
            ContextRestriction::Any => quote!(self.#method(context, command)),
            ContextRestriction::Player | ContextRestriction::Console => {
                quote!(self.#method(&mut restricted_context, command))
            }
        };
        let call = restricted_call(restriction, call);
        let index = command_types.len();
        command_types.push(quote!(#command_type));
        dispatch_arms.push(quote! {
            #index => match <#command_type as ::dragonfly_plugin::CommandDefinition>::parse(context.arguments()) {
                Ok(command) => { #call },
                Err(error) => {
                    context.fail(&error.to_string());
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
                    "a direct #[command] method must take &self and &mut Context",
                )
                .into_compile_error()
                .into();
            }
            let variant = syn::Ident::new(
                &format!("Variant{variant_index}"),
                proc_macro2::Span::call_site(),
            );
            let method = function.sig.ident.clone();
            let restriction = context_restriction(&function);
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
            let variant_attribute = if subcommand.is_empty() {
                quote!(#[command(root)])
            } else {
                quote!(#[command(name = #subcommand)])
            };
            variants.push(quote! {
                #variant_attribute
                #variant { #(#fields),* }
            });
            let call = match restriction {
                ContextRestriction::Any => quote!(self.#method(context, #(#field_names),*)),
                ContextRestriction::Player | ContextRestriction::Console => {
                    quote!(self.#method(&mut restricted_context, #(#field_names),*))
                }
            };
            let call = restricted_call(restriction, call);
            variant_dispatch.push(quote! {
                #command_type::#variant { #(#field_names),* } => { #call },
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
            #index => match <#command_type as ::dragonfly_plugin::CommandDefinition>::parse(context.arguments()) {
                Ok(command) => match command {
                    #(#variant_dispatch)*
                },
                Err(error) => {
                    context.fail(&error.to_string());
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
            fn on_command(&self, command: usize, context: &mut ::dragonfly_plugin::Context<'_>) {
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
    let handles_join = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_join"));
    let handles_quit = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_quit"));
    let handles_hurt = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_hurt"));
    let handles_heal = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_heal"));
    let handles_block_break = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_block_break"),
    );
    let handles_block_place = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_block_place"),
    );
    let handles_food_loss = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_food_loss"),
    );
    let handles_death = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_death"),
    );
    let handles_start_break = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_start_break"),
    );
    let handles_fire_extinguish = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_fire_extinguish"),
    );
    let handles_toggle_sprint = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_toggle_sprint"),
    );
    let handles_toggle_sneak = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_toggle_sneak"),
    );
    let handles_jump = implementation
        .items
        .iter()
        .any(|item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_jump"));
    let handles_teleport = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_teleport"),
    );
    let handles_experience_gain = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_experience_gain"),
    );
    let handles_punch_air = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_punch_air"),
    );
    let handles_held_slot_change = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_held_slot_change"),
    );
    let handles_sleep = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_sleep"),
    );
    let handles_block_pick = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_block_pick"),
    );
    let handles_lectern_page_turn = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_lectern_page_turn"),
    );
    let handles_sign_edit = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_sign_edit"),
    );
    let handles_item_use = implementation.items.iter().any(
        |item| matches!(item, syn::ImplItem::Fn(function) if function.sig.ident == "on_item_use"),
    );
    let subscriptions = u64::from(handles_move)
        | (u64::from(handles_chat) << 1)
        | (u64::from(handles_join) << 2)
        | (u64::from(handles_quit) << 3)
        | (u64::from(handles_hurt) << 4)
        | (u64::from(handles_heal) << 5)
        | (u64::from(handles_block_break) << 6)
        | (u64::from(handles_block_place) << 7)
        | (u64::from(handles_food_loss) << 8)
        | (u64::from(handles_death) << 9)
        | (u64::from(handles_start_break) << 10)
        | (u64::from(handles_fire_extinguish) << 11)
        | (u64::from(handles_toggle_sprint) << 12)
        | (u64::from(handles_toggle_sneak) << 13)
        | (u64::from(handles_jump) << 14)
        | (u64::from(handles_teleport) << 15)
        | (u64::from(handles_experience_gain) << 16)
        | (u64::from(handles_punch_air) << 17)
        | (u64::from(handles_held_slot_change) << 18)
        | (u64::from(handles_sleep) << 19)
        | (u64::from(handles_block_pick) << 20)
        | (u64::from(handles_lectern_page_turn) << 21)
        | (u64::from(handles_sign_edit) << 22)
        | (u64::from(handles_item_use) << 23);

    quote! {
        #[doc(hidden)]
        extern crate dragonfly as dragonfly_plugin;

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
                    let mut context = unsafe { ::dragonfly_plugin::Context::from_raw(&*input, &mut *state) };
                    <PluginType as ::dragonfly_plugin::Plugin>::on_command(plugin, command as usize, &mut context);
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
                    sys::DF_EVENT_PLAYER_JOIN => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerJoinInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerJoinState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerJoinEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_join(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_QUIT => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerQuitInput>() };
                        let event = unsafe { ::dragonfly_plugin::PlayerQuitEvent::from_raw(input) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_quit(plugin, &event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_HURT => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerHurtInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerHurtState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerHurtEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_hurt(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_HEAL => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerHealInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerHealState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerHealEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_heal(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_BLOCK_BREAK => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerBlockBreakInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerBlockBreakState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerBlockBreakEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_block_break(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_BLOCK_PLACE => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerBlockPlaceInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerBlockPlaceState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerBlockPlaceEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_block_place(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_FOOD_LOSS => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerFoodLossInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerFoodLossState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerFoodLossEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_food_loss(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_DEATH => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerDeathInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerDeathState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerDeathEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_death(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_START_BREAK => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerStartBreakInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerStartBreakState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerStartBreakEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_start_break(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_FIRE_EXTINGUISH => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerFireExtinguishInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerFireExtinguishState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerFireExtinguishEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_fire_extinguish(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_TOGGLE_SPRINT => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerToggleSprintInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerToggleSprintState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerToggleSprintEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_toggle_sprint(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_TOGGLE_SNEAK => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerToggleSneakInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerToggleSneakState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerToggleSneakEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_toggle_sneak(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_JUMP => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerJumpInput>() };
                        let event = unsafe { ::dragonfly_plugin::PlayerJumpEvent::from_raw(input) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_jump(plugin, &event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_TELEPORT => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerTeleportInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerTeleportState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerTeleportEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_teleport(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_EXPERIENCE_GAIN => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerExperienceGainInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerExperienceGainState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerExperienceGainEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_experience_gain(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_PUNCH_AIR => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerPunchAirInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerPunchAirState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerPunchAirEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_punch_air(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_HELD_SLOT_CHANGE => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerHeldSlotChangeInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerHeldSlotChangeState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerHeldSlotChangeEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_held_slot_change(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_SLEEP => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerSleepInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerSleepState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerSleepEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_sleep(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_BLOCK_PICK => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerBlockPickInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerBlockPickState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerBlockPickEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_block_pick(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_LECTERN_PAGE_TURN => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerLecternPageTurnInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerLecternPageTurnState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerLecternPageTurnEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_lectern_page_turn(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_SIGN_EDIT => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerSignEditInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerSignEditState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerSignEditEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_sign_edit(plugin, &mut event);
                        sys::DF_STATUS_OK
                    }
                    sys::DF_EVENT_PLAYER_ITEM_USE => {
                        let plugin = unsafe { &*instance.cast::<PluginType>() };
                        let input = unsafe { &*input.cast::<sys::DfPlayerItemUseInput>() };
                        let state = unsafe { &mut *state.cast::<sys::DfPlayerItemUseState>() };
                        let mut event = unsafe { ::dragonfly_plugin::PlayerItemUseEvent::from_raw(input, state) };
                        <PluginType as ::dragonfly_plugin::Plugin>::on_item_use(plugin, &mut event);
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
