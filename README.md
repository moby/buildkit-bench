# BuildKit Benchmarks

[![CI Status](https://img.shields.io/github/actions/workflow/status/moby/buildkit-bench/ci.yml?label=ci&logo=github&style=flat-square)](https://github.com/moby/buildkit-bench/actions?query=workflow%3Aci)

This repository contains a set of benchmarks for [BuildKit](https://github.com/moby/buildkit)
that run on GitHub Actions [public runners](https://github.com/actions/runner-images).
Results are published on [GitHub Pages](https://moby.github.io/buildkit-bench/).

![](.github/buildkit-bench.png)

___

* [Usage](#usage)
* [License](#license)

## Usage

To run locally, you can use the following command:

```bash
make test
```

This runs all tests and benchmarks from [./test](./test) package with BuildKit
changes from default branch. You can also specify a commit to test or multiple
references and tweak the benchmark settings:

```bash
# run a specific benchmark
TEST_BENCH_REGEXP=/BenchmarkBuildLocal$ make test

# run all benchmarks 3 times (default 1)
TEST_BENCH_RUN=3 make test

# run 5 iterations of each benchmark (default 1x)
TEST_BENCH_TIME=5x make test

# run all with master, v0.9.3 and v0.16.0 git references
BUILDKIT_REFS=master,v0.9.3,v0.16.0 make test
```

> [!NOTE]
> Set `TEST_KEEP_CACHE=1` for the test framework to keep external dependant
> images in a docker volume if you are repeatedly calling `./hack/test` script.
> This helps to avoid rate limiting on the remote registry side.

After running the tests, you can generate the HTML report and serve the
website with:

```bash
make gen
```

```text
...
[+] Running 1/0
 ✔ Container website-serve-1  Created                                                                                                                                                                     0.0s 
Attaching to serve-1
serve-1  | Running generateResults.js...
serve-1  | /src/public/results.json has been generated successfully.
serve-1  |  INFO  Starting development server...
 DONE  Compiled successfully in 947ms2:39:29 PM
serve-1  |
                                                                                                                                                                                                               
serve-1  |   App running at:
serve-1  |   - Local:   http://localhost:8080/
serve-1  |
serve-1  |   It seems you are running Vue CLI inside a container.
serve-1  |   Access the dev server via http://localhost:<your container's external mapped port>/
serve-1  |
serve-1  |   Note that the development build is not optimized.
serve-1  |   To create a production build, run yarn build.
serve-1  |
Build finished at 14:39:29 by 0.000s
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## License

This project is licensed under the Apache License, Version 2.0 - see the
[LICENSE](LICENSE) file for details.
