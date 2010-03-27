package stringset_test

import(
    "testing";
    "utilz/stringset";
)

func TestStringSet(t *testing.T){
    ss := stringset.New();
    ss.Add("en");
    if ss.Len() != 1 {
        t.FailNow();
    }
}
