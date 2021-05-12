import "github.com/weberc2/infra"

infra.constraints

#goBuild: infra.#GoBuild & {
    needs: ["comments-service:lint", "comments-service:test"]
}

projects: [
    {
        #pullRequestJobs: [ #goBuild ]
        #mergeJobs: [ #goBuild ]
        name: "comments-service"
    } & infra.#GoProject
]