package Component_test

import (
	"github.com/zllangct/RockGO/3rd/assert"
	c "github.com/zllangct/RockGO/component"
	"testing"
)

func TestBasicTemplateToObject(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		factory := c.NewObjectFactory()
		factory.Register(&FakeComponent{})
		template := c.ObjectTemplate{
			Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}},
			Objects: []c.ObjectTemplate{
				{Name: "First Child"},
				{Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}},
					Objects: []c.ObjectTemplate{
						{Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}}},
						{Name: "Last Child"}}}}}

		instance, err := factory.Deserialize(&template)

		T.Assert(err == nil)
		T.Assert(instance != nil)

		T.Assert(instance.Debug() == `object: Untitled (2 / 1)
! *github.com/zllangct/RockGO/Component_test.FakeComponent
   object: First Child (0 / 0)
   object: Untitled (2 / 1)
   ! *github.com/zllangct/RockGO/Component_test.FakeComponent
        object: Untitled (0 / 1)
        ! *github.com/zllangct/RockGO/Component_test.FakeComponent
        object: Last Child (0 / 0)
`)
	})
}

func TestComponentDeserialization(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		factory := c.NewObjectFactory()
		factory.Register(&FakeComponent{})
		template := c.ObjectTemplate{
			Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}},
			Objects: []c.ObjectTemplate{
				{Name: "First Child"},
				{Name: "D", Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent", Data: "Value2,5"}},
					Objects: []c.ObjectTemplate{
						{Name: "C", Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent", Data: "Value1,1"}}},
						{Name: "Last Child"}}}}}

		instance, err := factory.Deserialize(&template)

		T.Assert(err == nil)
		T.Assert(instance != nil)

		var c1 *FakeComponent
		var c2 *FakeComponent
		err = instance.Find(&c1, "D")
		err = instance.Find(&c2, "D", "C")

		T.Assert(c1.Id == "Value2")
		T.Assert(c1.Count == 5)
		T.Assert(c2.Id == "Value1")
		T.Assert(c2.Count == 1)
	})
}

func TestObjectToTemplate(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		factory := c.NewObjectFactory()
		factory.Register(&FakeComponent{})

		template := &c.ObjectTemplate{
			Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}},
			Objects: []c.ObjectTemplate{
				{Name: "First Child"},
				{Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}},
					Objects: []c.ObjectTemplate{
						{Components: []c.ComponentTemplate{{Type: "*github.com/zllangct/RockGO/Component_test.FakeComponent"}}},
						{Name: "Last Child"}}}}}

		instance, _ := factory.Deserialize(template)
		dump1 := instance.Debug()

		template, err := factory.Serialize(instance)
		T.Assert(err == nil)

		instance, err = factory.Deserialize(template)
		T.Assert(err == nil)
		T.Assert(instance != nil)

		dump2 := instance.Debug()
		T.Assert(dump1 == dump2)
	})
}

func TestTemplateToObjectViaFactory(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		factory := c.NewObjectFactory()
		factory.Register(&FakeComponent{})
		factory.Register(&FakeConfiguredComponent{})

		template, err := c.ObjectTemplateFromJson(objectTemplateSimple)
		T.Assert(err == nil)
		T.Assert(template != nil)

		instance, err := factory.Deserialize(template)
		T.Assert(err == nil)
		T.Assert(instance != nil)

		var cmp *FakeConfiguredComponent
		err = instance.Find(&cmp)
		T.Assert(err == nil)

		T.Assert(len(cmp.Data.Items) == 3)
		T.Assert(cmp.Data.Items[2].Id == "3")
		T.Assert(cmp.Data.Items[2].Count == 3)
	})
}

func TestNestedTemplateToObjectViaFactory(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		factory := c.NewObjectFactory()
		factory.Register(&FakeComponent{})
		factory.Register(&FakeConfiguredComponent{})

		template, err := c.ObjectTemplateFromJson(objectTemplateNested)
		T.Assert(err == nil)
		T.Assert(template != nil)

		instance, err := factory.Deserialize(template)
		T.Assert(err == nil)
		T.Assert(instance != nil)

		var cmp *FakeConfiguredComponent
		err = instance.Find(&cmp, "One", "Two")
		T.Assert(err == nil)

		T.Assert(len(cmp.Data.Items) == 3)
		T.Assert(cmp.Data.Items[2].Id == "3")
		T.Assert(cmp.Data.Items[2].Count == 3)
	})
}
