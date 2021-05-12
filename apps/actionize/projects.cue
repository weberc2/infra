import "github.com/weberc2/infra"

infra.constraints

projects: [
    {name: "actionize"} & infra.#GoProject
]