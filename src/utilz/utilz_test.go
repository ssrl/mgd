package utilz_test

import(
    "testing";
    "utilz/stringset";
    "utilz/stringbuffer";
)

func TestStringSet(t *testing.T){

    ss := stringset.New();

    ss.Add("en");

    if ss.Len() != 1 {
        t.Fatal("stringset.Len() != 1\n");
    }

    ss.Add("to");

    if ss.Len() != 2 {
        t.Fatal("stringset.Len() != 2\n");
    }

    if ! ss.Contains("en") {
        t.Fatal("! stringset.Contains('en')\n");
    }

    if ! ss.Contains("to") {
        t.Fatal("! stringset.Contains('to')\n");
    }

    if ss.Contains("not here") {
        t.Fatal(" stringset.Contains('not here')");
    }
}

func TestStringBuffer(t *testing.T){
    ss := stringbuffer.New();
    ss.Add("en");
    if ss.String() != "en" {
        t.FailNow();
    }
}

func BenchmarkStringBuffer(b *testing.B){
    b.StartTimer();
    cnt := 0;
    for i:= 0; i < 100; i++{
        cnt++;
    }
    b.StopTimer();
}

