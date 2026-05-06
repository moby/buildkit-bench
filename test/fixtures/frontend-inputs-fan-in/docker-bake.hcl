group "default" {
  targets = ["p19"]
}

target "p0" {
  context = "."
  dockerfile = "Dockerfile.p0"
}

target "p1" {
  context = "."
  dockerfile = "Dockerfile.p1"
  contexts = {
    ref_p0 = "target:p0"
  }
}

target "p2" {
  context = "."
  dockerfile = "Dockerfile.p2"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
  }
}

target "p3" {
  context = "."
  dockerfile = "Dockerfile.p3"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
  }
}

target "p4" {
  context = "."
  dockerfile = "Dockerfile.p4"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
  }
}

target "p5" {
  context = "."
  dockerfile = "Dockerfile.p5"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
  }
}

target "p6" {
  context = "."
  dockerfile = "Dockerfile.p6"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
  }
}

target "p7" {
  context = "."
  dockerfile = "Dockerfile.p7"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
  }
}

target "p8" {
  context = "."
  dockerfile = "Dockerfile.p8"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
  }
}

target "p9" {
  context = "."
  dockerfile = "Dockerfile.p9"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
  }
}

target "p10" {
  context = "."
  dockerfile = "Dockerfile.p10"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
  }
}

target "p11" {
  context = "."
  dockerfile = "Dockerfile.p11"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
  }
}

target "p12" {
  context = "."
  dockerfile = "Dockerfile.p12"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
  }
}

target "p13" {
  context = "."
  dockerfile = "Dockerfile.p13"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
  }
}

target "p14" {
  context = "."
  dockerfile = "Dockerfile.p14"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
    ref_p13 = "target:p13"
  }
}

target "p15" {
  context = "."
  dockerfile = "Dockerfile.p15"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
    ref_p13 = "target:p13"
    ref_p14 = "target:p14"
  }
}

target "p16" {
  context = "."
  dockerfile = "Dockerfile.p16"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
    ref_p13 = "target:p13"
    ref_p14 = "target:p14"
    ref_p15 = "target:p15"
  }
}

target "p17" {
  context = "."
  dockerfile = "Dockerfile.p17"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
    ref_p13 = "target:p13"
    ref_p14 = "target:p14"
    ref_p15 = "target:p15"
    ref_p16 = "target:p16"
  }
}

target "p18" {
  context = "."
  dockerfile = "Dockerfile.p18"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
    ref_p13 = "target:p13"
    ref_p14 = "target:p14"
    ref_p15 = "target:p15"
    ref_p16 = "target:p16"
    ref_p17 = "target:p17"
  }
}

target "p19" {
  context = "."
  dockerfile = "Dockerfile.p19"
  contexts = {
    ref_p0 = "target:p0"
    ref_p1 = "target:p1"
    ref_p2 = "target:p2"
    ref_p3 = "target:p3"
    ref_p4 = "target:p4"
    ref_p5 = "target:p5"
    ref_p6 = "target:p6"
    ref_p7 = "target:p7"
    ref_p8 = "target:p8"
    ref_p9 = "target:p9"
    ref_p10 = "target:p10"
    ref_p11 = "target:p11"
    ref_p12 = "target:p12"
    ref_p13 = "target:p13"
    ref_p14 = "target:p14"
    ref_p15 = "target:p15"
    ref_p16 = "target:p16"
    ref_p17 = "target:p17"
    ref_p18 = "target:p18"
  }
}
