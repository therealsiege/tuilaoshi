package main

// AllCharacters returns the combined list of all character data.
func AllCharacters() []Char {
	var all []Char
	all = append(all, HSK1Chars()...)
	all = append(all, HSK2Chars()...)
	all = append(all, HSK3Chars()...)
	all = append(all, HSK4Chars()...)
	all = append(all, HSK5Chars()...)
	all = append(all, HSK6Chars()...)
	// Extension batches
	all = append(all, HSK2ExtChars()...)
	all = append(all, HSK3ExtChars()...)
	all = append(all, HSK4Ext1Chars()...)
	all = append(all, HSK4Ext2Chars()...)
	all = append(all, HSK5Ext1Chars()...)
	all = append(all, HSK5Ext2Chars()...)
	all = append(all, HSK5Ext3Chars()...)
	all = append(all, HSK6Ext1Chars()...)
	all = append(all, HSK6Ext2Chars()...)
	all = append(all, HSK6Ext3Chars()...)
	all = append(all, HSK6Ext4Chars()...)
	all = append(all, HSK6Ext5Chars()...)
	all = append(all, HSK3CharsB()...)
	all = append(all, HSK6Ext1Part2Chars()...)
	all = append(all, HSK4CharsB()...)
	all = append(all, HSK5CharsB()...)
	all = append(all, HSK6CharsB()...)
	// Wave 3 — 3000 more
	all = append(all, HSK1ExtChars()...)
	all = append(all, HSK2Ext2Chars()...)
	all = append(all, HSK3Ext2Chars()...)
	all = append(all, HSK3Ext3Chars()...)
	all = append(all, HSK4Ext3Chars()...)
	all = append(all, HSK4Ext4Chars()...)
	all = append(all, HSK5Ext4Chars()...)
	all = append(all, HSK5Ext5Chars()...)
	all = append(all, HSK5Ext6Chars()...)
	all = append(all, HSK6Ext6Chars()...)
	all = append(all, HSK6Ext7Chars()...)
	all = append(all, HSK6Ext8Chars()...)
	all = append(all, HSK6Ext9Chars()...)
	all = append(all, HSK6Ext10Chars()...)
	all = append(all, HSK6Ext11Chars()...)
	// Domain packs
	all = append(all, BusinessChars()...)
	all = append(all, TechnologyChars()...)
	all = append(all, HealthcareChars()...)
	all = append(all, SportsChars()...)
	all = append(all, ChengyuChars()...)
	return all
}
