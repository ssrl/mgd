package stringbuffer_test

import(
    "testing";
    "utilz/stringbuffer";
)

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

