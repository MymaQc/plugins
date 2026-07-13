use super::{Command, CommandOverload, CommandParameter, CommandParameterKind, CommandValue};

const ENUM_VALUES: &[CommandValue] = &[CommandValue::new("one"), CommandValue::new("two")];
const PARAMETERS: &[CommandParameter] = &[
    CommandParameter::subcommand("action"),
    CommandParameter::enumeration("choice", ENUM_VALUES),
    CommandParameter::string("text"),
    CommandParameter::integer("count"),
    CommandParameter::float("distance"),
    CommandParameter::boolean("enabled"),
    CommandParameter::dynamic_enum("online"),
    CommandParameter::player("target").optional(),
    CommandParameter::raw_text("message"),
];
const OVERLOADS: &[CommandOverload] =
    &[CommandOverload::new(PARAMETERS), CommandOverload::new(&[])];
const COMMAND: Command = Command::new("sample", "Sample command").with_overloads(OVERLOADS);

#[test]
fn command_exposes_descriptor_metadata() {
    assert_eq!(COMMAND.name(), "sample");
    assert_eq!(COMMAND.description(), "Sample command");
    assert_eq!(COMMAND.overloads().len(), 2);

    let parameters = COMMAND.overloads()[0].parameters();
    assert_eq!(
        parameters
            .iter()
            .map(CommandParameter::kind)
            .collect::<Vec<_>>(),
        vec![
            CommandParameterKind::Subcommand,
            CommandParameterKind::Enum,
            CommandParameterKind::String,
            CommandParameterKind::Integer,
            CommandParameterKind::Float,
            CommandParameterKind::Boolean,
            CommandParameterKind::DynamicEnum,
            CommandParameterKind::Player,
            CommandParameterKind::RawText,
        ]
    );
    assert_eq!(parameters[7].name(), "target");
    assert!(parameters[7].is_optional());
    assert!(!parameters[6].is_optional());
    assert_eq!(
        parameters[1]
            .values()
            .iter()
            .map(CommandValue::as_str)
            .collect::<Vec<_>>(),
        vec!["one", "two"]
    );
}

#[test]
fn empty_descriptor_collections_are_empty_slices() {
    const EMPTY: Command = Command::new("empty", "Empty command");

    assert!(EMPTY.overloads().is_empty());
    assert!(COMMAND.overloads()[1].parameters().is_empty());
    assert!(PARAMETERS[0].values().is_empty());
}
