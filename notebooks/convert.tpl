#!/usr/bin/env ipython
import time
{% extends 'python.tpl'%}

{% block codecell %}
{% if "kubectl rollout status" in super() or "delete" in super() %}
time.sleep(10)
{{ super() }}
time.sleep(5)
{% elif "docker run" in super() %}
{{ super() }}
time.sleep(3)
{% elif "kubectl wait" in super() %}
{{ super() }}
time.sleep(5)
{% else %}
{{ super().replace("'inline'","'agg'") }}
{% endif %}
{% endblock codecell %}