import "github.com/weberc2/infra"

infra.constraints

#name: "comments-service"

#goBuild: infra.#GoBuild & {
    needs: ["\(#name):lint", "\(#name):test"]
}

projects: [
    {
        #pullRequestJobs: [ #goBuild ]
        #mergeJobs: [ #goBuild ]
        name: #name
    } & infra.#GoProject
]