package gengo

import (
	"testing"

	. "github.com/warpfork/go-wish"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/must"
	"github.com/ipld/go-ipld-prime/schema"
)

func TestMapsContainingMaybe(t *testing.T) {
	ts := schema.TypeSystem{}
	ts.Init()
	adjCfg := &AdjunctCfg{
		maybeUsesPtr: map[schema.TypeName]bool{},
	}
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnMap("Map__String__String",
		ts.TypeByName("String"), ts.TypeByName("String"), false))
	ts.Accumulate(schema.SpawnMap("Map__String__nullableString",
		ts.TypeByName("String"), ts.TypeByName("String"), true))

	test := func(t *testing.T, getStyleByName func(string) ipld.NodeStyle) {
		t.Run("non-nullable", func(t *testing.T) {
			ns := getStyleByName("Map__String__String")
			nsr := getStyleByName("Map__String__String.Repr")
			var n schema.TypedNode
			t.Run("typed-create", func(t *testing.T) {
				n = fluent.MustBuildMap(ns, 2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("one").AssignString("1")
					ma.AssembleEntry("two").AssignString("2")
				}).(schema.TypedNode)
				t.Run("typed-read", func(t *testing.T) {
					Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
					Wish(t, n.Length(), ShouldEqual, 2)
					Wish(t, must.String(must.Node(n.LookupString("one"))), ShouldEqual, "1")
					Wish(t, must.String(must.Node(n.LookupString("two"))), ShouldEqual, "2")
					_, err := n.LookupString("miss")
					Wish(t, err, ShouldBeSameTypeAs, ipld.ErrNotExists{})
				})
				t.Run("repr-read", func(t *testing.T) {
					nr := n.Representation()
					Require(t, nr.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
					Wish(t, nr.Length(), ShouldEqual, 2)
					Wish(t, must.String(must.Node(nr.LookupString("one"))), ShouldEqual, "1")
					Wish(t, must.String(must.Node(nr.LookupString("two"))), ShouldEqual, "2")
					_, err := nr.LookupString("miss")
					Wish(t, err, ShouldBeSameTypeAs, ipld.ErrNotExists{})
				})
			})
			t.Run("repr-create", func(t *testing.T) {
				nr := fluent.MustBuildMap(nsr, 2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("one").AssignString("1")
					ma.AssembleEntry("two").AssignString("2")
				})
				Wish(t, n, ShouldEqual, nr)
			})
		})
		t.Run("nullable", func(t *testing.T) {
			ns := getStyleByName("Map__String__nullableString")
			nsr := getStyleByName("Map__String__nullableString.Repr")
			var n schema.TypedNode
			t.Run("typed-create", func(t *testing.T) {
				n = fluent.MustBuildMap(ns, 2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("one").AssignString("1")
					ma.AssembleEntry("none").AssignNull()
				}).(schema.TypedNode)
				t.Run("typed-read", func(t *testing.T) {
					Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
					Wish(t, n.Length(), ShouldEqual, 2)
					Wish(t, must.String(must.Node(n.LookupString("one"))), ShouldEqual, "1")
					Wish(t, must.Node(n.LookupString("none")), ShouldEqual, ipld.Null)
					_, err := n.LookupString("miss")
					Wish(t, err, ShouldBeSameTypeAs, ipld.ErrNotExists{})
				})
				t.Run("repr-read", func(t *testing.T) {
					nr := n.Representation()
					Require(t, nr.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
					Wish(t, nr.Length(), ShouldEqual, 2)
					Wish(t, must.String(must.Node(nr.LookupString("one"))), ShouldEqual, "1")
					Wish(t, must.Node(nr.LookupString("none")), ShouldEqual, ipld.Null)
					_, err := nr.LookupString("miss")
					Wish(t, err, ShouldBeSameTypeAs, ipld.ErrNotExists{})
				})
			})
			t.Run("repr-create", func(t *testing.T) {
				nr := fluent.MustBuildMap(nsr, 2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("one").AssignString("1")
					ma.AssembleEntry("none").AssignNull()
				})
				Wish(t, n, ShouldEqual, nr)
			})
		})
	}

	t.Run("maybe-using-embed", func(t *testing.T) {
		adjCfg.maybeUsesPtr["String"] = false

		prefix := "maps-embed"
		pkgName := "main"
		genAndCompileAndTest(t, prefix, pkgName, ts, adjCfg, func(t *testing.T, getStyleByName func(string) ipld.NodeStyle) {
			test(t, getStyleByName)
		})
	})
	t.Run("maybe-using-ptr", func(t *testing.T) {
		adjCfg.maybeUsesPtr["String"] = true

		prefix := "maps-mptr"
		pkgName := "main"
		genAndCompileAndTest(t, prefix, pkgName, ts, adjCfg, func(t *testing.T, getStyleByName func(string) ipld.NodeStyle) {
			test(t, getStyleByName)
		})
	})
}

// TestMapsContainingMaps is probing *two* things:
//   - that maps can nest, obviously
//   - that representation semantics are held correctly when we recurse, both in builders and in reading
// To cover that latter situation, this depends on structs (so we can use rename directives on the representation to make it distinctive).
func TestMapsContainingMaps(t *testing.T) {
	ts := schema.TypeSystem{}
	ts.Init()
	adjCfg := &AdjunctCfg{
		maybeUsesPtr: map[schema.TypeName]bool{},
	}
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnStruct("Frub", // "type Frub struct { field String (rename "encoded") }"
		[]schema.StructField{
			schema.SpawnStructField("field", ts.TypeByName("String"), false, false), // plain field.
		},
		schema.SpawnStructRepresentationMap(map[string]string{
			"field": "encoded",
		}),
	))
	ts.Accumulate(schema.SpawnMap("Map__String__Frub", // "{String:Frub}"
		ts.TypeByName("String"), ts.TypeByName("Frub"), false))
	ts.Accumulate(schema.SpawnMap("Map__String__nullableMap__String__Frub", // "{String:nullable {String:Frub}}"
		ts.TypeByName("String"), ts.TypeByName("Map__String__Frub"), true))

	prefix := "maps-recursive"
	pkgName := "main"
	genAndCompileAndTest(t, prefix, pkgName, ts, adjCfg, func(t *testing.T, getStyleByName func(string) ipld.NodeStyle) {
		ns := getStyleByName("Map__String__nullableMap__String__Frub")
		nsr := getStyleByName("Map__String__nullableMap__String__Frub.Repr")
		creation := func(t *testing.T, ns ipld.NodeStyle, fieldName string) schema.TypedNode {
			return fluent.MustBuildMap(ns, 3, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("one").CreateMap(2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("zot").CreateMap(1, func(ma fluent.MapAssembler) { ma.AssembleEntry(fieldName).AssignString("11") })
					ma.AssembleEntry("zop").CreateMap(1, func(ma fluent.MapAssembler) { ma.AssembleEntry(fieldName).AssignString("12") })
				})
				ma.AssembleEntry("two").CreateMap(1, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("zim").CreateMap(1, func(ma fluent.MapAssembler) { ma.AssembleEntry(fieldName).AssignString("21") })
				})
				ma.AssembleEntry("none").AssignNull()
			}).(schema.TypedNode)
		}
		reading := func(t *testing.T, n ipld.Node, fieldName string) {
			withNode(n, func(n ipld.Node) {
				Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
				Wish(t, n.Length(), ShouldEqual, 3)
				withNode(must.Node(n.LookupString("one")), func(n ipld.Node) {
					Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
					Wish(t, n.Length(), ShouldEqual, 2)
					withNode(must.Node(n.LookupString("zot")), func(n ipld.Node) {
						Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
						Wish(t, n.Length(), ShouldEqual, 1)
						Wish(t, must.String(must.Node(n.LookupString(fieldName))), ShouldEqual, "11")
					})
					withNode(must.Node(n.LookupString("zop")), func(n ipld.Node) {
						Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
						Wish(t, n.Length(), ShouldEqual, 1)
						Wish(t, must.String(must.Node(n.LookupString(fieldName))), ShouldEqual, "12")
					})
				})
				withNode(must.Node(n.LookupString("two")), func(n ipld.Node) {
					Wish(t, n.Length(), ShouldEqual, 1)
					withNode(must.Node(n.LookupString("zim")), func(n ipld.Node) {
						Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
						Wish(t, n.Length(), ShouldEqual, 1)
						Wish(t, must.String(must.Node(n.LookupString(fieldName))), ShouldEqual, "21")
					})
				})
				withNode(must.Node(n.LookupString("none")), func(n ipld.Node) {
					Wish(t, n, ShouldEqual, ipld.Null)
				})
				_, err := n.LookupString("miss")
				Wish(t, err, ShouldBeSameTypeAs, ipld.ErrNotExists{})
			})
		}
		var n schema.TypedNode
		t.Run("typed-create", func(t *testing.T) {
			n = creation(t, ns, "field")
			t.Run("typed-read", func(t *testing.T) {
				reading(t, n, "field")
			})
			t.Run("repr-read", func(t *testing.T) {
				reading(t, n.Representation(), "encoded")
			})
		})
		t.Run("repr-create", func(t *testing.T) {
			nr := creation(t, nsr, "encoded")
			Wish(t, n, ShouldEqual, nr)
		})
	})
}

func TestMapsWithComplexKeys(t *testing.T) {
	ts := schema.TypeSystem{}
	ts.Init()
	adjCfg := &AdjunctCfg{
		maybeUsesPtr: map[schema.TypeName]bool{},
	}
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnStruct("StringyStruct",
		[]schema.StructField{
			schema.SpawnStructField("foo", ts.TypeByName("String"), false, false),
			schema.SpawnStructField("bar", ts.TypeByName("String"), false, false),
		},
		schema.SpawnStructRepresentationStringjoin(":"),
	))
	ts.Accumulate(schema.SpawnMap("Map__StringyStruct__String",
		ts.TypeByName("StringyStruct"), ts.TypeByName("String"), false))

	prefix := "maps-cmplx-keys"
	pkgName := "main"
	genAndCompileAndTest(t, prefix, pkgName, ts, adjCfg, func(t *testing.T, getStyleByName func(string) ipld.NodeStyle) {
		ns := getStyleByName("Map__StringyStruct__String")
		nsr := getStyleByName("Map__StringyStruct__String.Repr")
		var n schema.TypedNode
		t.Run("typed-create", func(t *testing.T) {
			n = fluent.MustBuildMap(ns, 3, func(ma fluent.MapAssembler) {
				ma.AssembleKey().CreateMap(2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("foo").AssignString("a")
					ma.AssembleEntry("bar").AssignString("b")
				})
				ma.AssembleValue().AssignString("1")
				ma.AssembleKey().CreateMap(2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("foo").AssignString("c")
					ma.AssembleEntry("bar").AssignString("d")
				})
				ma.AssembleValue().AssignString("2")
				ma.AssembleKey().CreateMap(2, func(ma fluent.MapAssembler) {
					ma.AssembleEntry("foo").AssignString("e")
					ma.AssembleEntry("bar").AssignString("f")
				})
				ma.AssembleValue().AssignString("3")
			}).(schema.TypedNode)
			t.Run("typed-read", func(t *testing.T) {
				Require(t, n.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
				Wish(t, n.Length(), ShouldEqual, 3)
				n2 := must.Node(n.LookupString("c:d"))
				Require(t, n2.ReprKind(), ShouldEqual, ipld.ReprKind_String)
				Wish(t, must.String(n2), ShouldEqual, "2")
			})
			t.Run("repr-read", func(t *testing.T) {
				nr := n.Representation()
				Require(t, nr.ReprKind(), ShouldEqual, ipld.ReprKind_Map)
				Wish(t, nr.Length(), ShouldEqual, 3)
				n2 := must.Node(nr.LookupString("c:d"))
				Require(t, n2.ReprKind(), ShouldEqual, ipld.ReprKind_String)
				Wish(t, must.String(n2), ShouldEqual, "2")
			})
		})
		t.Run("repr-create", func(t *testing.T) {
			nr := fluent.MustBuildMap(nsr, 3, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("a:b").AssignString("1")
				ma.AssembleEntry("c:d").AssignString("2")
				ma.AssembleEntry("e:f").AssignString("3")
			})
			Wish(t, n, ShouldEqual, nr)
		})
	})
}
