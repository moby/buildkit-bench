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
changes from default branch and latest Buildx stable. You can also specify a
commit to test or multiple references and tweak the benchmark settings:

```bash
# run only tests
TEST_TYPES=test make test

# run only benchmarks
TEST_TYPES=benchmark make test
# or
make bench

# run a specific benchmark
TEST_BENCH_REGEXP=/BenchmarkBuildLocal$ make bench

# run all benchmarks 3 times (default 1)
TEST_BENCH_RUN=3 make bench

# run 5 iterations of each benchmark (default 1x)
TEST_BENCH_TIME=5x make bench

# run all with master, v0.9.3 and v0.16.0 buildkit references (with latest buildx stable)
BUILDKIT_REFS=master,v0.9.3,v0.16.0 make bench

# run all with master, v0.17.0 buildx git references (with latest buildkit stable)
BUILDX_REFS=master,v0.17.0 make bench
```

> [!NOTE]
> Set `TEST_KEEP_CACHE=1` for the test framework to keep external dependant
> images in a docker volume if you are repeatedly calling `make test` or
> `make bench`. This helps to avoid rate limiting on the remote registry side.

After running the tests, you can generate the HTML report and serve the
website with:

```bash
make gen
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## License

This project is licensed under the Apache License, Version 2.0 - see the
[LICENSE](LICENSE) file for details.
