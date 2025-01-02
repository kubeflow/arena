#
# Copyright 2024 The Kubeflow Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# 	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

import argparse
import os

import torch
import torch.distributed as dist
import torch.nn.functional as F
from torch import nn, optim
from torch.optim.lr_scheduler import StepLR
from torch.utils.tensorboard import SummaryWriter
from torchvision import datasets, transforms


class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 32, 3, 1)
        self.conv2 = nn.Conv2d(32, 64, 3, 1)
        self.dropout1 = nn.Dropout(0.25)
        self.dropout2 = nn.Dropout(0.5)
        self.fc1 = nn.Linear(9216, 128)
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        x = self.conv1(x)
        x = F.relu(x)
        x = self.conv2(x)
        x = F.relu(x)
        x = F.max_pool2d(x, 2)
        x = self.dropout1(x)
        x = torch.flatten(x, 1)
        x = self.fc1(x)
        x = F.relu(x)
        x = self.dropout2(x)
        x = self.fc2(x)
        output = F.log_softmax(x, dim=1)
        return output


def train(args, model, device, train_loader, optimizer, epoch, writer):
    model.train()
    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(device), target.to(device)
        optimizer.zero_grad()
        output = model(data)
        loss = F.nll_loss(output, target)
        loss.backward()
        optimizer.step()
        if batch_idx % args.log_interval == 0:
            print(
                "Train Epoch: {} [{}/{} ({:.0f}%)]\tLoss: {:.6f}".format(
                    epoch,
                    batch_idx * len(data),
                    len(train_loader.dataset),
                    100.0 * batch_idx / len(train_loader),
                    loss.item(),
                )
            )
            niter = epoch * len(train_loader) + batch_idx
            writer.add_scalar('loss', loss.item(), niter)
            if args.dry_run:
                break


def test(model, device, test_loader, epoch, writer):
    model.eval()
    test_loss = 0
    correct = 0
    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            # sum up batch loss
            test_loss += F.nll_loss(output, target, reduction="sum").item()
            # get the index of the max log-probability
            pred = output.argmax(dim=1, keepdim=True)
            correct += pred.eq(target.view_as(pred)).sum().item()

    test_loss /= len(test_loader.dataset)
    accuracy = float(correct) / len(test_loader.dataset)
    print(
        "\nAccuracy: {}/{} ({:.2f}%)\n".format(
            correct,
            len(test_loader.dataset),
            accuracy * 100.0,
        )
    )
    writer.add_scalar('accuracy', accuracy, epoch)


def print_env():
    info = {
        "PID": os.getpid(),
        "MASTER_ADDR": os.environ["MASTER_ADDR"],
        "MASTER_PORT": os.environ["MASTER_PORT"],
        "LOCAL_RANK": int(os.environ["LOCAL_RANK"]),
        "RANK": int(os.environ["RANK"]),
        "GROUP_RANK": int(os.environ["GROUP_RANK"]),
        "ROLE_RANK": int(os.environ["ROLE_RANK"]),
        "LOCAL_WORLD_SIZE": int(os.environ["LOCAL_WORLD_SIZE"]),
        "WORLD_SIZE": int(os.environ["WORLD_SIZE"]),
        "ROLE_WORLD_SIZE": int(os.environ["ROLE_WORLD_SIZE"]),
    }
    print(info)


def main():
    parser = argparse.ArgumentParser(description="PyTorch MNIST Example")
    parser.add_argument(
        "--data",
        default="../data",
        metavar="D",
        help="directory where summary logs are stored",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=64,
        metavar="N",
        help="input batch size for training (default: 64)",
    )
    parser.add_argument(
        "--test-batch-size",
        type=int,
        default=1000,
        metavar="N",
        help="input batch size for testing (default: 1000)",
    )
    parser.add_argument(
        "--epochs",
        type=int,
        default=14,
        metavar="N",
        help="number of epochs to train (default: 14)",
    )
    parser.add_argument(
        "--lr",
        type=float,
        default=1.0,
        metavar="LR",
        help="learning rate (default: 1.0)",
    )
    parser.add_argument(
        "--gamma",
        type=float,
        default=0.7,
        metavar="M",
        help="Learning rate step gamma (default: 0.7)",
    )
    parser.add_argument(
        "--no-cuda",
        action="store_true",
        default=False,
        help="disables CUDA training",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        default=False,
        help="quickly check a single pass",
    )
    parser.add_argument(
        "--seed",
        type=int,
        default=1,
        metavar="S",
        help="random seed (default: 1)"
    )
    parser.add_argument(
        "--log-interval",
        type=int,
        default=10,
        metavar="N",
        help="how many batches to wait before logging training status",
    )
    parser.add_argument(
        "--save-model",
        action="store_true",
        default=False,
        help="For Saving the current Model",
    )
    parser.add_argument(
        '--dir',
        default=os.path.join(os.path.dirname(__file__), 'logs'),
        metavar='L',
        help='directory where summary logs are stored'
    )
    if dist.is_available():
        parser.add_argument(
            "--backend",
            type=str,
            default=dist.Backend.NCCL,
            choices=[
                dist.Backend.NCCL,
                dist.Backend.GLOO,
                dist.Backend.MPI
            ],
            help="Distributed backend",
        )
    args = parser.parse_args()
    print_env()

    torch.manual_seed(args.seed)
    use_cuda = not args.no_cuda and torch.cuda.is_available()

    train_kwargs = {"batch_size": args.batch_size}
    test_kwargs = {"batch_size": args.test_batch_size}
    if use_cuda:
        cuda_kwargs = {
            "num_workers": 1,
            "pin_memory": True,
            "shuffle": True
        }
        train_kwargs.update(cuda_kwargs)
        test_kwargs.update(cuda_kwargs)
    transform = transforms.Compose([
        transforms.ToTensor(),
        transforms.Normalize((0.1307,), (0.3081,))
    ])
    train_dataset = datasets.MNIST(
        args.data,
        train=True,
        download=True,
        transform=transform
    )
    test_dataset = datasets.MNIST(
        args.data,
        train=False,
        transform=transform
    )
    train_loader = torch.utils.data.DataLoader(
        train_dataset,
        **train_kwargs
    )
    test_loader = torch.utils.data.DataLoader(
        test_dataset,
        **test_kwargs
    )

    if use_cuda:
        device_id = int(os.environ["LOCAL_RANK"])
        print(f"Using cuda:{device_id}.")
        device = torch.device(f"cuda:{device_id}")
    else:
        print("Using cpu")
        device = torch.device("cpu")
    model = Net().to(device)

    world_size = int(os.environ["WORLD_SIZE"])
    is_distributed = dist.is_available() and world_size > 1
    if is_distributed:
        dist.init_process_group(args.backend)
        model = nn.parallel.DistributedDataParallel(model)

    optimizer = optim.Adadelta(model.parameters(), lr=args.lr)
    scheduler = StepLR(optimizer, step_size=1, gamma=args.gamma)

    writer = SummaryWriter(args.dir)
    for epoch in range(1, args.epochs + 1):
        train(args, model, device, train_loader, optimizer, epoch, writer)
        test(model, device, test_loader, epoch, writer)
        scheduler.step()

    if args.save_model:
        torch.save(model.state_dict(), "mnist_cnn.pt")

    if is_distributed:
        dist.destroy_process_group()


if __name__ == "__main__":
    main()
