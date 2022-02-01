package logging

// Structured logging keys.
//
// Practitioners or tooling reading logs may be depending on these keys, so be
// conscious of that when changing them.
//
// Refer to the terraform-plugin-go logging keys as well, which should be
// equivalent to these when possible.
const (
	// The type of resource being operated on, such as "random_pet"
	KeyResourceType = "tf_resource_type"

	// The type of data source being operated on, such as "archive_file"
	KeyDataSourceType = "tf_data_source_type"
)
