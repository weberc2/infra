import "github.com/weberc2/infra"

projects: [
    {name: "bootstrap-tf"} & infra.#TerraformProject
]