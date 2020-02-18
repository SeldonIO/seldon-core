#!/usr/bin/env ipython
import time
{% extends 'python.tpl'%}

{% block codecell %}
{% if "kubectl rollout status" in super() or "delete" in super() %}
{{ super() }}
time.sleep(4)
{% else %}
{{ super() }}
{% endif %}
{% endblock codecell %}