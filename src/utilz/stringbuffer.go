// © Knug Industries 2010 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package stringbuffer


type StringBuffer struct{
    current, max int;
    buffer []byte;
}


func New() *StringBuffer{
    s := new(StringBuffer);
    s.Clear();
    return s;
}

func NewSize(size int) *StringBuffer{
    s := new(StringBuffer);
    s.current = 0;
    s.max = size;
    s.buffer = make([]byte, size);
    return s;
}

func (s *StringBuffer) Add(more string){

    if ( len(more) + s.current ) > s.max {
        s.resize();
        s.Add(more);

    }else{

        var lenmore int = len(more);

        for i := 0; i < lenmore; i++ {
            s.buffer[i + s.current] = more[i];
        }

        s.current += lenmore;
    }
}

func (s *StringBuffer) Clear(){
    s.buffer = make([]byte, 100);
    s.current = 0;
    s.max = 100;
}

func (s *StringBuffer) ClearSize(z int){
    s.buffer = make([]byte, z);
    s.current = 0;
    s.max = z;
}

func (s *StringBuffer) Capacity() int{
    return s.max;
}

func (s *StringBuffer) Len() int{
    return s.current;
}

func (s *StringBuffer) String() string{
    slice := s.buffer[0:s.current];
    return string(slice);
}

func (s *StringBuffer) resize(){
    s.max = s.max * 2;
    nbuffer := make([]byte, s.max);

    for i := 0; i < s.current; i++ {
        nbuffer[i] = s.buffer[i];
    }

    s.buffer = nbuffer;
}