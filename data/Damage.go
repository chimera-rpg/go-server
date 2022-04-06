package data

type Damage struct {
	Value          int
	AttributeBonus map[AttackType]map[AttributeType]float64
}

func (d *Damage) Add(other Damage) {
	d.Value += other.Value
	for k, v := range other.AttributeBonus {
		if _, ok := d.AttributeBonus[k]; !ok {
			d.AttributeBonus[k] = v
		} else {
			for k2, v2 := range v {
				if _, ok := d.AttributeBonus[k][k2]; !ok {
					d.AttributeBonus[k][k2] = v2
				} else {
					d.AttributeBonus[k][k2] += v2
				}
			}
		}
	}
}
