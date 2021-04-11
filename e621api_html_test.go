package e621api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func tags(t ...string) []string {
	return t
}

func Test_BlacklistEntry_NoNegatives(t *testing.T) {
	be := createBlacklistEntry(tags("gay", "fox"))

	assert.Equal(t, true, be.matches(tags("gay", "fox", "white")), "superset")
	assert.Equal(t, false, be.matches(tags("gay", "white")), "intersect")
	assert.Equal(t, true, be.matches(tags("gay", "fox")), "exact")
	assert.Equal(t, false, be.matches(tags("gay")), "subset")
	assert.Equal(t, false, be.matches(tags()), "empty")
}

func Test_BlacklistEntry_PositiveAndNegative(t *testing.T) {
	be := createBlacklistEntry(tags("gay", "fox", "-wolf", "-lion"))

	assert.Equal(t, true, be.matches(tags("gay", "fox", "white")), "superset")
	assert.Equal(t, false, be.matches(tags("gay", "white")), "intersect")
	assert.Equal(t, true, be.matches(tags("gay", "fox")), "exact")
	assert.Equal(t, false, be.matches(tags("gay")), "subset")
	assert.Equal(t, false, be.matches(tags()), "empty")

	assert.Equal(t, false, be.matches(tags("gay", "fox", "white", "wolf")), "superset+negative")
	assert.Equal(t, false, be.matches(tags("gay", "white", "wolf")), "intersect+negative")
	assert.Equal(t, false, be.matches(tags("gay", "fox", "wolf")), "exact+negative")
	assert.Equal(t, false, be.matches(tags("gay", "wolf")), "subset+negative")
}

func Test_BlacklistEntry_OnlyNegatives(t *testing.T) {
	be := createBlacklistEntry(tags("-gay"))

	assert.Equal(t, false, be.matches(tags("gay", "fox")), "contained")
	assert.Equal(t, true, be.matches(tags("plain", "fox")), "missing")
	assert.Equal(t, true, be.matches(tags()), "empty")
}

func Test_GetDefaultBlacklist(t *testing.T) {
	actual, err := CreateE621Api().GetDefaultBlacklist()
	if err != nil {
		t.Fatal(err)
		return
	}

	expected := []BlacklistEntry{
		createBlacklistEntry(tags("gore")),
		createBlacklistEntry(tags("scat")),
		createBlacklistEntry(tags("watersports")),
		createBlacklistEntry(tags("young", "-rating:s")),
		createBlacklistEntry(tags("loli")),
		createBlacklistEntry(tags("shota")),
	}

	assert.Equal(t, actual, expected, "")
}
