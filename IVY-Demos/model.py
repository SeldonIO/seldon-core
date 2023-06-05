import torch
import torch.nn as nn
import torch.optim as optim
import ivy

def mse_loss(y_pred, y_true):
    return ivy.mean(ivy.square(y_pred - y_true))

class MLP(ivy.Module):
    def __init__(self, input_size, hidden_size, output_size):
        super(MLP, self).__init__()
        
        self.fc1 = ivy.Linear(input_size, hidden_size)
        self.fc2 = ivy.Linear(hidden_size, output_size)
        
    def _forward(self, x):
        x = ivy.tanh(self.fc1(x))
        x = ivy.sqrt(self.fc2(x))
        return x

# Create the MLP model
mlp = MLP(input_size=1, hidden_size=10, output_size=1)

# Define the loss function and optimizer in ivy
criterion = mse_loss
optimizer = ivy.SGD(lr=0.01)

# Generate some random training data
X_train = ivy.random.randint(0, 100, shape=[1000, 1])  # Input numbers
y_train = ivy.sqrt(X_train)   # Square roots of the input numbers

# Train the model
for epoch in range(1000):
    optimizer.
    outputs = mlp(X_train)
    loss = criterion(outputs, y_train)
    loss.backward()
    optimizer.step()
    if epoch % 100 == 0:
        print(f"Epoch {epoch}/{1000}, Loss: {loss.item()}")

# Test the model
X_test = ivy.array([[4.0], [9.0], [16.0]])  # Test numbers
predictions = mlp(X_test)
print(f"Predictions: {predictions}")