package util

import (
  "encoding/json"
  "encoding/xml"
  "fmt"
  "github.com/peter-mount/golib/codec"
  "strconv"
)

// Public Timetable time
// Note: 00:00 is not possible as in CIF that means no-time
type PublicTime struct {
  t int
}

func (a *PublicTime) Equals( b *PublicTime ) bool {
  if b == nil {
    return false
  }
  if a.IsZero() {
    return b.IsZero()
  }
  return a.t == b.t
}

// Compare a PublicTime against another, accounting for crossing midnight.
// The rules for handling crossing midnight are:
// < -6 hours = crossed midnight
// < 0 back in time
// < 18 hours increasing time
// > 18 hours back in time & crossing midnight
func (a *PublicTime) Compare( b *PublicTime ) bool {
  if b == nil {
    return false
  }

  d := a.t - b.t

  if d < -360 || d >1080 {
    return d > 0
  }

  return d < 0
}

// PublicTimeEquals compares equality between two PublicTime.
// Unlike PublicTime.Equals() this will return true if both are null, otherwise
// both must not be null and equal to be true

func PublicTimeEquals( a *PublicTime, b *PublicTime ) bool {
  return ( a == nil && b == nil ) ||
         ( a != nil && a.Equals( b ) )
}

// BinaryCodec writer
func (t *PublicTime) Write( c *codec.BinaryCodec ) {
  if t.IsZero() {
    // If zero then write a single byte value 255
    c.WriteByte( byte(255) )
  } else {
    // 16 bit number, Big endian so we should not hit 255,xxx in the stream
    c.WriteByte( byte( t.t >> 8 ) ).
      WriteByte( byte( t.t ) )
  }
}

// BinaryCodec reader
func (t *PublicTime) Read( c *codec.BinaryCodec ) {
  var b byte

  c.ReadByte( &b )
  if b == byte(255) {
    t.t = -1
  } else {
    var i uint64 = uint64(b)<<8
    c.ReadByte( &b )
    t.t = int( i | uint64(b) )
  }
}

// NewPublicTime returns a new PublicTime instance from a string of format "HH:MM"
func NewPublicTime( s string ) *PublicTime {
  v := &PublicTime{}
  v.Parse( s )
  return v
}

func (v *PublicTime) Parse( s string ) {
  if s == "" {
    v.t = -1
  } else {
    a, _ := strconv.Atoi( s[0:2] )
    b, _ := strconv.Atoi( s[3:5] )
    v.Set( (a *60) + b )
  }
}

// PublicTimeRead is a workaround issue where a custom type cannot be
// omitempty in JSON unless it's a nil
// So instead of using BinaryCodec.Read( v ), we call this & set the return
// value in the struct as a pointer.
func PublicTimeRead( c *codec.BinaryCodec ) *PublicTime {
  t := &PublicTime{}
  c.Read( t )
  if t.IsZero() {
    return nil
  }
  return t
}

// PublicTimeWrite is a workaround for writing null times.
// If the pointer is null then a time is written where IsZero()==true
func PublicTimeWrite( c *codec.BinaryCodec, t *PublicTime ) {
  if t == nil {
    // Null so mark as zero
    c.WriteByte( byte(255) )
  } else {
    c.Write( t )
  }
}

// Custom JSON Marshaler. This will write null or the time as "HH:MM"
func (t *PublicTime) MarshalJSON() ( []byte, error ) {
  if t.IsZero() {
    return json.Marshal( nil )
  }
  return json.Marshal( t.String() )
}

func (t *PublicTime) UnmarshalJSON( b []byte ) error {
  s := string(b[:])
  if s != "null" && len( s ) > 2 {
    t.Parse( s[1:len(s)-1] )
  }
  return nil
}

// Custom XML Marshaler.
func (t *PublicTime) MarshalXMLAttr( name xml.Name ) ( xml.Attr, error ) {
  if t.IsZero() {
    return xml.Attr{}, nil
  }
  return xml.Attr{ Name: name, Value: t.String() }, nil
}

// String returns a PublicTime in HH:MM format or 5 blank spaces if it's not set.
func (t *PublicTime) String() string {
  if t.IsZero() {
    return "     "
  }

  return fmt.Sprintf( "%02d:%02d", t.t/60, t.t%60 )
}

// Get returns the PublicTime in minutes of the day
func (t *PublicTime) Get() int {
  return t.t
}

// Set sets the PublicTime in minutes of the day
func (t *PublicTime) Set( v int ) {
  t.t = v
}

// IsZero returns true if the time is not present
func (t *PublicTime) IsZero() bool {
  return t.t <= 0
}