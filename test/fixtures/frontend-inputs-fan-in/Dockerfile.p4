# fan-in fixture p4 padding 00 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 01 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 02 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 03 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 04 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 05 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 06 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 07 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 08 repeated frontend input source map data for gateway Inputs payload compression measurement.
# fan-in fixture p4 padding 09 repeated frontend input source map data for gateway Inputs payload compression measurement.
FROM busybox:latest AS stage
WORKDIR /work
COPY --from=ref_p0 /work/marker /work/marker_p0
COPY --from=ref_p1 /work/marker /work/marker_p1
COPY --from=ref_p2 /work/marker /work/marker_p2
COPY --from=ref_p3 /work/marker /work/marker_p3
RUN echo p4 > /work/marker
