import ivy

# Define the network
class MLP(ivy.Module):
    def __init__(self):
        super(MLP, self).__init__()
        self.linear1 = ivy.Linear(1, 10)
        self.relu = ivy.ReLU()
        self.linear2 = ivy.Linear(10, 1)

    def _forward(self, x):
        out = self.linear1(x)
        out = self.relu(out)
        out = self.linear2(out)
        return out
    

