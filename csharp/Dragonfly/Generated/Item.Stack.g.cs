// Code generated from Dragonfly server/item/stack.go Go AST. DO NOT EDIT.
#nullable enable
using System.Collections.Generic;

namespace Dragonfly;

public static partial class Item
{
    public static Stack NewStack(World.Item t, int count) =>
        NewStackImpl(t, count);

    public readonly partial struct Stack
    {
        public int Count() =>
            CountImpl();

        public int MaxCount() =>
            MaxCountImpl();

        public Stack Grow(int n) =>
            GrowImpl(n);

        public int Durability() =>
            DurabilityImpl();

        public int MaxDurability() =>
            MaxDurabilityImpl();

        public Stack Damage(int d) =>
            DamageImpl(d);

        public Stack WithDurability(int d) =>
            WithDurabilityImpl(d);

        public bool Unbreakable() =>
            UnbreakableImpl();

        public Stack AsUnbreakable() =>
            AsUnbreakableImpl();

        public Stack AsBreakable() =>
            AsBreakableImpl();

        public bool Empty() =>
            EmptyImpl();

        public World.Item? Item() =>
            ItemImpl();

        public double AttackDamage() =>
            AttackDamageImpl();

        public Stack WithCustomName(params object?[] a) =>
            WithCustomNameImpl(a);

        public string CustomName() =>
            CustomNameImpl();

        public Stack WithLore(params string[] lines) =>
            WithLoreImpl(lines);

        public string[] Lore() =>
            LoreImpl();

        public Stack WithValue(string key, object? val) =>
            WithValueImpl(key, val);

        public (object? Value, bool Ok) Value(string key) =>
            ValueImpl(key);

        public Stack WithEnchantments(params Enchantment[] enchants) =>
            WithEnchantmentsImpl(enchants);

        public Stack WithForcedEnchantments(params Enchantment[] enchants) =>
            WithForcedEnchantmentsImpl(enchants);

        public Stack WithoutEnchantments(params EnchantmentType[] enchants) =>
            WithoutEnchantmentsImpl(enchants);

        public (Enchantment Enchantment, bool Ok) Enchantment(EnchantmentType enchant) =>
            EnchantmentImpl(enchant);

        public Enchantment[] Enchantments() =>
            EnchantmentsImpl();

        public int AnvilCost() =>
            AnvilCostImpl();

        public Stack WithAnvilCost(int anvilCost) =>
            WithAnvilCostImpl(anvilCost);

        public Stack WithItem(World.Item t) =>
            WithItemImpl(t);

        public (Stack A, Stack B) AddStack(Stack s2) =>
            AddStackImpl(s2);

        public bool Equal(Stack s2) =>
            EqualImpl(s2);

        public bool Comparable(Stack s2) =>
            ComparableImpl(s2);

        public string String() =>
            StringImpl();

        public IReadOnlyDictionary<string, object> Values() =>
            ValuesImpl();
    }
}
