/^---/ {
  x;
  /name: .*hodometer/ {
    i \{\{ if .Values.hodometer.enabled \}\}
    a \{\{ end \}\}
  };
  1!p;
  d
};
H;
${
  g;
  p;
}

