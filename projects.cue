import "github.com/weberc2/infra"

// Import projects constraints
infra.constraints

projects: [
    // Make sure `actionize` has been run on the repo by running it and making
    // sure nothing has changed.
    {
        name: "actionizecheck"
        jobs: {
            "pull-request": [
                infra.#Job & {
                    name: "check"
                    #steps: [
                        infra.#CheckoutStep,
                        infra.#GoSetupStep,
                        {
                            name: "check"
                            run: """
                                set -eo pipefail
                                cd {{ .Path }}/apps/actionize
                                go run .
                                if [[ -n "$(git diff)" ]]; then
                                    echo "Found differences:"
                                    echo "$(git diff)"
                                    exit 1
                                fi
                                """
                        },
                    ]
                }
            ]
        }
    }
]