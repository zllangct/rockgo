package Component_test

var objectTemplateSimple = `
{
	"Name": "Sample Object",
	"Components": [
		{"Type": "*github.com/zllangct/RockGO/Component_test.FakeComponent"},
		{"Type": "*github.com/zllangct/RockGO/Component_test.FakeConfiguredComponent", "Data": {
			"Items": [
				{"Id": "1", "Count": 1},
				{"Id": "2", "Count": 2},
				{"Id": "3", "Count": 3}
			]
		}}
	],
	"Objects": [
	]
}
`

var objectTemplateNested = `
{
	"Name": "Sample Object",
	"Components": [{
		"Type": "*github.com/zllangct/RockGO/Component_test.FakeComponent"
	}],
	"Objects": [{
			"Name": "N/A",
			"Components": [{
				"Type": "*github.com/zllangct/RockGO/Component_test.FakeComponent"
			}],
			"Objects": []
		}, {
			"Name": "One",
			"Objects": [{
				"Name": "Two",
				"Components": [{
					"Type": "*github.com/zllangct/RockGO/Component_test.FakeComponent"
				}, {
					"Type": "*github.com/zllangct/RockGO/Component_test.FakeConfiguredComponent",
					"Data": {
						"Items": [{
							"Id": "1",
							"Count": 1
						}, {
							"Id": "2",
							"Count": 2
						}, {
							"Id": "3",
							"Count": 3
						}]
					}
				}],
				"Objects": []
			}]
		}
	]
}
`
