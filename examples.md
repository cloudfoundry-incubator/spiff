```
spiff merge a.yml b.yml c.yml > abc.yml

a.yml
------------------------------
hello:
  world
happy:
  (( merge ))

b.yml
------------------------------
foo:
  bar

c.yml
------------------------------
happy:
  birthday

abc.yml
==============================
happy: birthday
hello: world

*Note: happy will reach all the way down to c to grab “birthday”
```

```
spiff merge a.yml b.yml > ab.yml

a.yml
------------------------------
hello:
  world
happy: ~

b.yml
------------------------------
foo:
  bar
happy:
  new_year

ab.yml
==============================
happy: new_year
hello: world

*Note: When the “happy” key appears in two places, the right most value wins out. There is an implicit merge.
```

```
spiff merge a.yml b.yml > ab.yml
a.yml
------------------------------
happy:
  - name: holidays
    date: yesterday
  - name: new_year
    date: tomorrow

b.yml
------------------------------
happy:
  - name: birthday
    date: today
  - name: new_year
    date: "12/31"

ab.yml
==============================
happy:
- date: yesterday
  name: holidays
- date: 12/31
  name: new_year

* Arrays, are a special exception. “name” is a special key, which
identifies a member of an array for possible merging. If there is overlap of
name during a merge, the rightmost file wins, for that element.
```