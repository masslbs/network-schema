type write struct {
    patchs: []patch
}

type patch struct {
    op: string
    path: string
    // ?? 
    value: map[string]interface{}
}
