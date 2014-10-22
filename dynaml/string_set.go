package dynaml

type StringSet map[string]bool

func (self StringSet) Add(k string) {
	self[k] = true
}

func (self StringSet) Has(k string) bool {
	v, ok := self[k]
	return v && ok
}

func (self StringSet) UpdateSlice(other []string) {
	for _, k := range other {
		self.Add(k)
	}
}

func (self StringSet) Update(other StringSet) {
	for k := range other {
		self.Add(k)
	}
}

func (self StringSet) Copy() StringSet {
	retval := StringSet{}
	for k := range self {
		retval[k] = true
	}
	return retval
}

func (self StringSet) Len() int {
	return len(self)
}

func (self StringSet) Union(other StringSet) StringSet {
	retval := self.Copy()
	retval.Update(other)
	return retval
}

func (self StringSet) Difference(other StringSet) StringSet {
	retval := StringSet{}
	for k := range self {
		v, ok := other[k]
		if !ok || !v {
			retval.Add(k)
		}
	}
	return retval
}
