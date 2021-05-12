import "github.com/weberc2/infra"

infra.constraints

projects: [
    {
        #pullRequestJobs: [ infra.#GoBuild ]
        #mergeJobs: [ infra.#GoBuild ]
        name: "comments-service"
    } & infra.#GoProject
]