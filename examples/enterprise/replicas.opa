lint[output] {
  spec := specs[_]
  spec.spec.replicas == 1
  field := concat(".", [spec.field, "replicas"])
  output := {
    "type": "error",
    "message": "Cannot accept hard coded replica count when it's set to 1",
    "path": spec.path,
    "field": field,
    "docIndex": spec.docIndex
  }
}
