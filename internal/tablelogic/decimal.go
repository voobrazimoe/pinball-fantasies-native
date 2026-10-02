package tablelogic

import "fmt"

// Decimal is the original twelve unpacked decimal bytes, most significant first.
// ADDSCOREBCD's final carry is discarded (modulo 10^12), including alias addition.
type Decimal [12]uint8

func Number(n uint64) (d Decimal) {
	for i := 11; i >= 0; i-- {
		d[i] = uint8(n % 10)
		n /= 10
	}
	return
}
func (d Decimal) Uint64() (n uint64) {
	for _, v := range d {
		n = n*10 + uint64(v)
	}
	return
}
func (d Decimal) String() string { return fmt.Sprintf("%012d", d.Uint64()) }
func (d *Decimal) Add(v Decimal) {
	carry := uint8(0)
	for i := 11; i >= 0; i-- {
		n := d[i] + v[i] + carry
		d[i] = n % 10
		carry = n / 10
	}
}
func (d *Decimal) AddNumber(n uint64) { d.Add(Number(n)) }
