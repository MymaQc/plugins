#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Food {
    quick_regeneration: bool,
}

impl Food {
    pub const fn new(quick_regeneration: bool) -> Self {
        Self { quick_regeneration }
    }

    pub const fn quick_regeneration(self) -> bool {
        self.quick_regeneration
    }
}

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub struct Custom<'a> {
    name: &'a str,
}

impl<'a> Custom<'a> {
    pub const fn new(name: &'a str) -> Self {
        Self { name }
    }

    pub const fn name(self) -> &'a str {
        self.name
    }
}

#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub struct Instant;

#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub struct Regeneration;

#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub enum Source<'a> {
    Food(Food),
    Instant,
    Regeneration,
    Custom(Custom<'a>),
}

impl<'a> Source<'a> {
    pub const fn name(&self) -> &str {
        match self {
            Self::Food(_) => "food",
            Self::Instant => "instant",
            Self::Regeneration => "regeneration",
            Self::Custom(custom) => custom.name(),
        }
    }

    pub(crate) fn with_raw<R>(
        &self,
        callback: impl FnOnce(&dragonfly_plugin_sys::DfHealingSourceView) -> R,
    ) -> R {
        let (kind, data) = match self {
            Self::Food(source) => (
                dragonfly_plugin_sys::DF_HEALING_SOURCE_FOOD,
                u8::from(source.quick_regeneration()),
            ),
            Self::Instant => (dragonfly_plugin_sys::DF_HEALING_SOURCE_INSTANT, 0),
            Self::Regeneration => (dragonfly_plugin_sys::DF_HEALING_SOURCE_REGENERATION, 0),
            Self::Custom(_) => (dragonfly_plugin_sys::DF_HEALING_SOURCE_CUSTOM, 0),
        };
        callback(&dragonfly_plugin_sys::DfHealingSourceView {
            name: crate::string_view_from_str(self.name()),
            kind,
            data,
        })
    }

    pub(crate) unsafe fn from_raw(
        raw: &'a dragonfly_plugin_sys::DfHealingSourceView,
    ) -> Option<Self> {
        Some(match raw.kind {
            dragonfly_plugin_sys::DF_HEALING_SOURCE_FOOD => Self::Food(Food::new(raw.data != 0)),
            dragonfly_plugin_sys::DF_HEALING_SOURCE_INSTANT => Self::Instant,
            dragonfly_plugin_sys::DF_HEALING_SOURCE_REGENERATION => Self::Regeneration,
            dragonfly_plugin_sys::DF_HEALING_SOURCE_CUSTOM => {
                Self::Custom(Custom::new(unsafe { crate::string_view(raw.name) }))
            }
            _ => return None,
        })
    }
}

impl<'a> From<Food> for Source<'a> {
    fn from(value: Food) -> Self {
        Self::Food(value)
    }
}

impl<'a> From<Instant> for Source<'a> {
    fn from(_: Instant) -> Self {
        Self::Instant
    }
}

impl<'a> From<Regeneration> for Source<'a> {
    fn from(_: Regeneration) -> Self {
        Self::Regeneration
    }
}

impl<'a> From<Custom<'a>> for Source<'a> {
    fn from(value: Custom<'a>) -> Self {
        Self::Custom(value)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn food_preserves_quick_regeneration() {
        let source: Source<'_> = Food::new(true).into();
        assert!(matches!(source, Source::Food(food) if food.quick_regeneration()));
        source.with_raw(|raw| {
            let decoded = unsafe { Source::from_raw(raw) }.unwrap();
            assert_eq!(decoded, source);
        });
    }

    #[test]
    fn unit_sources_convert_without_adapter_code() {
        assert_eq!(Source::from(Instant), Source::Instant);
        assert_eq!(Source::from(Regeneration), Source::Regeneration);
    }

    #[test]
    fn custom_name_borrows_the_callback_input() {
        let source: Source<'_> = Custom::new("example.CustomHealingSource").into();
        source.with_raw(|raw| {
            let decoded = unsafe { Source::from_raw(raw) }.unwrap();
            assert_eq!(decoded.name(), "example.CustomHealingSource");
        });
    }
}
