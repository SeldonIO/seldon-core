from bokeh.io import show, curdoc
from bokeh.layouts import row, column, widgetbox
from bokeh.models.widgets import Dropdown, Button, Select
from bokeh.plotting import figure, ColumnDataSource
from bokeh import palettes
from bokeh.models import GraphRenderer, Oval, StaticLayoutProvider


mabs_options = [
    ("Epsilon Greedy", "seldonio/mab_epsilon_greedy:0.6"), 
    ("Thomson Sampling", "seldonio/mab_thomson_sampling:0.6"), 
    ("Contextual Linear", "seldonio/mab_contextual_ts:0.6"), 
    ("Contextual Binary", "seldonio/mab_contextual_gbnb:0.6")]

model_options = [
    ("Simple","seldonio/mean_classifier:0.4"),
    ("Average","seldonio/mean_classifier:0.4")]

dropdown_mab = Select(title="MAB", options=mabs_options)
dropdown_model_1 = Select(title="Model 1", options=model_options)
dropdown_model_2 = Select(title="Model 2", options=model_options)
dropdown_model_3 = Select(title="Model 3", options=model_options)

button_deploy = Button(label="Deploy", button_type="success")
button_delete = Button(label="Delete", button_type="danger")

p = figure(plot_width=600, plot_height=600)
show(row(widgetbox(dropdown_mab,dropdown_model_1,dropdown_model_2,dropdown_model_3,button_deploy,button_delete),p))
