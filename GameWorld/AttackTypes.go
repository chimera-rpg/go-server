package GameWorld

type AttackTypes struct {
  physical int
  fire int
  cold int
}

/*
ModifyDamage reduces the pointed to AttackTypes' values by the values in the caller's AttackTypes.
*/
func (r AttackTypes) ModifyDamage(d *AttackTypes) {
  if d.physical != 0 {
    d.physical = d.physical - r.physical
  }
  if d.fire != 0 {
    d.fire = d.fire - r.fire
  }
  if d.cold != 0 {
    d.cold = d.cold - r.cold
  }
}
