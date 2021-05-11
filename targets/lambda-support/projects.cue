import "github.com/weberc2/infra"

projects: [
    {name: "lambda-support"} & infra.#TerraformProject
]