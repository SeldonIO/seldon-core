
# Tensorflow GPU MNIST Model

In this tutorial we will run a deep MNIST Tensorflow example with GPU.

The tutorial will be broken down into the following sections:

1. Install all dependencies to run Tensorflow-GPU
    
    1.1 Installing CUDA 9.0
    
    1.2 Installing cuDNN 7.0
    
    1.3 Configure CUDA and cuDNN
    
    1.4 Install Tensorflow GPU
    
    
2. Train the MNIST model locally


3. Test using minikube


## 1) Installing all dependencies to run Tensorflow-GPU

* Dependencies installed in this section:
    * Nvidia compute 3.0 onwards
    * CUDA 9.0
    * cuDNN 7.0
    * tensorflow-gpu 1.12.0

**Check Nvidia drivers >= 3.0**


```python
!nvidia-smi
```

    Wed Jun  5 15:38:27 2019       
    +-----------------------------------------------------------------------------+
    | NVIDIA-SMI 410.104      Driver Version: 410.104      CUDA Version: 10.0     |
    |-------------------------------+----------------------+----------------------+
    | GPU  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC |
    | Fan  Temp  Perf  Pwr:Usage/Cap|         Memory-Usage | GPU-Util  Compute M. |
    |===============================+======================+======================|
    |   0  Tesla V100-SXM2...  On   | 00000000:00:04.0 Off |                    0 |
    | N/A   34C    P0    25W / 300W |      0MiB / 16130MiB |      0%      Default |
    +-------------------------------+----------------------+----------------------+
                                                                                   
    +-----------------------------------------------------------------------------+
    | Processes:                                                       GPU Memory |
    |  GPU       PID   Type   Process name                             Usage      |
    |=============================================================================|
    |  No running processes found                                                 |
    +-----------------------------------------------------------------------------+


## 1.1) Install CUDA 9.0

* **Download the CUDA 9.0 runfile**


```python
!wget https://developer.nvidia.com/compute/cuda/9.0/Prod/local_installers/cuda_9.0.176_384.81_linux-run
```

    --2019-06-05 15:38:47--  https://developer.nvidia.com/compute/cuda/9.0/Prod/local_installers/cuda_9.0.176_384.81_linux-run
    Resolving developer.nvidia.com (developer.nvidia.com)... 192.229.162.216
    Connecting to developer.nvidia.com (developer.nvidia.com)|192.229.162.216|:443... connected.
    HTTP request sent, awaiting response... 302 Found
    Location: https://developer.download.nvidia.com/compute/cuda/9.0/secure/Prod/local_installers/cuda_9.0.176_384.81_linux.run?J_eX6Q6hlIw2Xo8Lxlemv90ViqlfeVUzqC-MyALWN5f26ddhQsDIcmt9wkZzkI9ouig00tob9MijUjUdv6sZGbiN1HLrTNNI7u2Hu6Nycd9fT-a9vVRotkDkL4GWXJvnzCF40IxZ-VKo_O2Amcsfh-XZIU0RZyrBdm63b49JuBxBXWSh_k88fMFq [following]
    --2019-06-05 15:38:47--  https://developer.download.nvidia.com/compute/cuda/9.0/secure/Prod/local_installers/cuda_9.0.176_384.81_linux.run?J_eX6Q6hlIw2Xo8Lxlemv90ViqlfeVUzqC-MyALWN5f26ddhQsDIcmt9wkZzkI9ouig00tob9MijUjUdv6sZGbiN1HLrTNNI7u2Hu6Nycd9fT-a9vVRotkDkL4GWXJvnzCF40IxZ-VKo_O2Amcsfh-XZIU0RZyrBdm63b49JuBxBXWSh_k88fMFq
    Resolving developer.download.nvidia.com (developer.download.nvidia.com)... 192.229.211.70, 2606:2800:21f:3aa:dcf:37b:1ed6:1fb
    Connecting to developer.download.nvidia.com (developer.download.nvidia.com)|192.229.211.70|:443... connected.
    HTTP request sent, awaiting response... 200 OK
    Length: 1643293725 (1.5G) [application/octet-stream]
    Saving to: ‘cuda_9.0.176_384.81_linux-run’
    
    cuda_9.0.176_384.81 100%[===================>]   1.53G   265MB/s    in 6.0s    
    
    2019-06-05 15:38:53 (261 MB/s) - ‘cuda_9.0.176_384.81_linux-run’ saved [1643293725/1643293725]
    


* **Unpack the separate files:**


```python
!chmod +x cuda_9.0.176_384.81_linux-run
! ./cuda_9.0.176_384.81_linux-run --extract=$HOME
```

    Logging to /tmp/cuda_install_18328.log
    Extracting individual Driver, Toolkit and Samples installers to /root ...


* **Install the Cuda 9.0 Toolkit file**:

From the terminal, run the following command
```
$ sudo ~/cuda-linux.9.0.176-22781540.run
```
Hold 'd' to scroll to the bottom of the license agreement.

Accept the licencing agreement and all of the default settings.

* **Verify the install, by installing the sample test:**
```
$ sudo ~/cuda-samples.9.0.176-22781540-linux.run
```
Again, accept the agreement and all default settings

* **Configure the runtime library:**

```
$ sudo bash -c "echo /usr/local/cuda/lib64/ > /etc/ld.so.conf.d/cuda.conf"
```

```
$ sudo ldconfig
```

* **Add the cuda bin to the file system:**

```
$ sudo vim /etc/environment
```

Add ‘:/usr/local/cuda/bin’ to the end of the PATH (inside quotes)

* **Reboot the system**


```python
!sudo shutdown -r now
```

* **Run the tests that we set up** - this takes some time to complete, so let it run for a little while...


```bash
%%bash

cd /usr/local/cuda-9.0/samples
sudo make
```

    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleIPC'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleIPC'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTemplates_nvrtc'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTemplates_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/inlinePTX_nvrtc'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/inlinePTX_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleMultiGPU'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleMultiGPU'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleLayeredTexture'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleLayeredTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simplePitchLinearTexture'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simplePitchLinearTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/vectorAddDrv'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/vectorAddDrv'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/asyncAPI'
    make[1]: Nothing to be done for 'all'.
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/asyncAPI'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/cudaOpenMP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cudaOpenMP.o -c cudaOpenMP.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cudaOpenMP cudaOpenMP.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp cudaOpenMP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/cudaOpenMP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTexture'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTexture.o -c simpleTexture.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTexture simpleTexture.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleTexture ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/inlinePTX'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o inlinePTX.o -c inlinePTX.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o inlinePTX inlinePTX.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp inlinePTX ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/inlinePTX'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAssert_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o simpleAssert.o -c simpleAssert.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o simpleAssert_nvrtc simpleAssert.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleAssert_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAssert_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTemplates'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTemplates.o -c simpleTemplates.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTemplates simpleTemplates.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleTemplates ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTemplates'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAtomicIntrinsics'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleAtomicIntrinsics.o -c simpleAtomicIntrinsics.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleAtomicIntrinsics_cpu.o -c simpleAtomicIntrinsics_cpu.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleAtomicIntrinsics simpleAtomicIntrinsics.o simpleAtomicIntrinsics_cpu.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleAtomicIntrinsics ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAtomicIntrinsics'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMul_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o matrixMul.o -c matrixMul.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o matrixMul_nvrtc matrixMul.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp matrixMul_nvrtc ../../bin/x86_64/linux/release
    cp ""/usr/local/cuda"/include/cooperative_groups.h" .
    cp ""/usr/local/cuda"/include/cooperative_groups_helpers.h" .
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMul_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/cudaTensorCoreGemm'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -maxrregcount=255 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cudaTensorCoreGemm.o -c cudaTensorCoreGemm.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cudaTensorCoreGemm cudaTensorCoreGemm.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp cudaTensorCoreGemm ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/cudaTensorCoreGemm'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleVoteIntrinsics_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o simpleVoteIntrinsics.o -c simpleVoteIntrinsics.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o simpleVoteIntrinsics_nvrtc simpleVoteIntrinsics.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleVoteIntrinsics_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleVoteIntrinsics_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMul'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o matrixMul.o -c matrixMul.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o matrixMul matrixMul.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp matrixMul ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMul'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/clock'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o clock.o -c clock.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o clock clock.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp clock ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/clock'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleStreams'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleStreams.o -c simpleStreams.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleStreams simpleStreams.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleStreams ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleStreams'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/systemWideAtomics'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o systemWideAtomics.o -c systemWideAtomics.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o systemWideAtomics systemWideAtomics.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp systemWideAtomics ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/systemWideAtomics'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleZeroCopy'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleZeroCopy.o -c simpleZeroCopy.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleZeroCopy simpleZeroCopy.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleZeroCopy ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleZeroCopy'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleSurfaceWrite'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleSurfaceWrite.o -c simpleSurfaceWrite.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleSurfaceWrite simpleSurfaceWrite.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleSurfaceWrite ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleSurfaceWrite'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simplePrintf'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simplePrintf.o -c simplePrintf.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simplePrintf simplePrintf.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simplePrintf ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simplePrintf'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleCubemapTexture'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCubemapTexture.o -c simpleCubemapTexture.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCubemapTexture simpleCubemapTexture.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCubemapTexture ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleCubemapTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTextureDrv'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o simpleTextureDrv.o -c simpleTextureDrv.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o simpleTextureDrv simpleTextureDrv.o  -lcuda
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleTextureDrv ../../bin/x86_64/linux/release
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o simpleTexture_kernel64.ptx -ptx simpleTexture_kernel.cu
    mkdir -p data
    cp -f simpleTexture_kernel64.ptx ./data
    mkdir -p ../../bin/x86_64/linux/release
    cp -f simpleTexture_kernel64.ptx ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleTextureDrv'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleMultiCopy'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleMultiCopy.o -c simpleMultiCopy.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleMultiCopy simpleMultiCopy.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleMultiCopy ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleMultiCopy'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/fp16ScalarProduct'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fp16ScalarProduct.o -c fp16ScalarProduct.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fp16ScalarProduct fp16ScalarProduct.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp fp16ScalarProduct ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/fp16ScalarProduct'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/clock_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o clock.o -c clock.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o clock_nvrtc clock.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp clock_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/clock_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/vectorAdd'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o vectorAdd.o -c vectorAdd.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o vectorAdd vectorAdd.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp vectorAdd ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/vectorAdd'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/UnifiedMemoryStreams'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o UnifiedMemoryStreams.o -c UnifiedMemoryStreams.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o UnifiedMemoryStreams UnifiedMemoryStreams.o  -lcublas
    mkdir -p ../../bin/x86_64/linux/release
    cp UnifiedMemoryStreams ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/UnifiedMemoryStreams'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/cdpSimplePrint'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpSimplePrint.o -c cdpSimplePrint.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpSimplePrint cdpSimplePrint.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp cdpSimplePrint ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/cdpSimplePrint'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleSeparateCompilation'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleDeviceLibrary.o -c simpleDeviceLibrary.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -lib -o simpleDeviceLibrary.a simpleDeviceLibrary.o 
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleSeparateCompilation.o -c simpleSeparateCompilation.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleSeparateCompilation simpleDeviceLibrary.a simpleSeparateCompilation.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleSeparateCompilation ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleSeparateCompilation'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/vectorAdd_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o vectorAdd.o -c vectorAdd.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o vectorAdd_nvrtc vectorAdd.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp vectorAdd_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/vectorAdd_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleVoteIntrinsics'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleVoteIntrinsics.o -c simpleVoteIntrinsics.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleVoteIntrinsics simpleVoteIntrinsics.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleVoteIntrinsics ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleVoteIntrinsics'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/cdpSimpleQuicksort'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpSimpleQuicksort.o -c cdpSimpleQuicksort.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpSimpleQuicksort cdpSimpleQuicksort.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp cdpSimpleQuicksort ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/cdpSimpleQuicksort'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleOccupancy'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleOccupancy.o -c simpleOccupancy.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleOccupancy simpleOccupancy.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleOccupancy ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleOccupancy'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAssert'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleAssert.o -c simpleAssert.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleAssert simpleAssert.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleAssert ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAssert'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAtomicIntrinsics_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o simpleAtomicIntrinsics.o -c simpleAtomicIntrinsics.cpp
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o simpleAtomicIntrinsics_cpu.o -c simpleAtomicIntrinsics_cpu.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o simpleAtomicIntrinsics_nvrtc simpleAtomicIntrinsics.o simpleAtomicIntrinsics_cpu.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleAtomicIntrinsics_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleAtomicIntrinsics_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleMPI'
    -----------------------------------------------------------------------------------------------
    WARNING - No MPI compiler found.
    -----------------------------------------------------------------------------------------------
    CUDA Sample "simpleMPI" cannot be built without an MPI Compiler.
    This will be a dry-run of the Makefile.
    For more information on how to set up your environment to build and run this 
    sample, please refer the CUDA Samples documentation and release notes
    -----------------------------------------------------------------------------------------------
    [@] mpicxx -I../../common/inc -o simpleMPI_mpi.o -c simpleMPI.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleMPI.o -c simpleMPI.cu
    [@] mpicxx -o simpleMPI simpleMPI_mpi.o simpleMPI.o -L/usr/local/cuda/lib64 -lcudart
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleMPI ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleMPI'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/template'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o template.o -c template.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o template_cpu.o -c template_cpu.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o template template.o template_cpu.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp template ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/template'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMulDrv'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o matrixMulDrv.o -c matrixMulDrv.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o matrixMulDrv matrixMulDrv.o  -lcuda
    mkdir -p ../../bin/x86_64/linux/release
    cp matrixMulDrv ../../bin/x86_64/linux/release
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o matrixMul_kernel64.ptx -ptx matrixMul_kernel.cu
    mkdir -p data
    cp -f matrixMul_kernel64.ptx ./data
    mkdir -p ../../bin/x86_64/linux/release
    cp -f matrixMul_kernel64.ptx ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMulDrv'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/cppIntegration'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cppIntegration.o -c cppIntegration.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cppIntegration_gold.o -c cppIntegration_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cppIntegration cppIntegration.o cppIntegration_gold.o main.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp cppIntegration ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/cppIntegration'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/cppOverload'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cppOverload.o -c cppOverload.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cppOverload cppOverload.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp cppOverload ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/cppOverload'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMulCUBLAS'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o matrixMulCUBLAS.o -c matrixMulCUBLAS.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o matrixMulCUBLAS matrixMulCUBLAS.o  -lcublas
    mkdir -p ../../bin/x86_64/linux/release
    cp matrixMulCUBLAS ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/matrixMulCUBLAS'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleP2P'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleP2P.o -c simpleP2P.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleP2P simpleP2P.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleP2P ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleP2P'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleCooperativeGroups'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCooperativeGroups.o -c simpleCooperativeGroups.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCooperativeGroups simpleCooperativeGroups.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCooperativeGroups ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleCooperativeGroups'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/0_Simple/simpleCallback'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o multithreading.o -c multithreading.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCallback.o -c simpleCallback.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCallback multithreading.o simpleCallback.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCallback ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/0_Simple/simpleCallback'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/1_Utilities/topologyQuery'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o topologyQuery.o -c topologyQuery.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o topologyQuery topologyQuery.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp topologyQuery ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/1_Utilities/topologyQuery'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/1_Utilities/p2pBandwidthLatencyTest'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o p2pBandwidthLatencyTest.o -c p2pBandwidthLatencyTest.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o p2pBandwidthLatencyTest p2pBandwidthLatencyTest.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp p2pBandwidthLatencyTest ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/1_Utilities/p2pBandwidthLatencyTest'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/1_Utilities/bandwidthTest'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bandwidthTest.o -c bandwidthTest.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bandwidthTest bandwidthTest.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp bandwidthTest ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/1_Utilities/bandwidthTest'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/1_Utilities/deviceQuery'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o deviceQuery.o -c deviceQuery.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o deviceQuery deviceQuery.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp deviceQuery ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/1_Utilities/deviceQuery'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/1_Utilities/deviceQueryDrv'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o deviceQueryDrv.o -c deviceQueryDrv.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o deviceQueryDrv deviceQueryDrv.o  -lcuda
    mkdir -p ../../bin/x86_64/linux/release
    cp deviceQueryDrv ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/1_Utilities/deviceQueryDrv'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/volumeFiltering'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volume.o -c volume.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeFilter_kernel.o -c volumeFilter_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeFiltering.o -c volumeFiltering.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeRender_kernel.o -c volumeRender_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeFiltering volume.o volumeFilter_kernel.o volumeFiltering.o volumeRender_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp volumeFiltering ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/volumeFiltering'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/Mandelbrot'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o Mandelbrot.o -c Mandelbrot.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o Mandelbrot_cuda.o -c Mandelbrot_cuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o Mandelbrot_gold.o -c Mandelbrot_gold.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o Mandelbrot Mandelbrot.o Mandelbrot_cuda.o Mandelbrot_gold.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp Mandelbrot ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/Mandelbrot'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGL'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGL.o -c simpleGL.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGL simpleGL.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleGL ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGL'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/bindlessTexture'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bindlessTexture.o -c bindlessTexture.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bindlessTexture_kernel.o -c bindlessTexture_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bindlessTexture bindlessTexture.o bindlessTexture_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp bindlessTexture ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/bindlessTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/volumeRender'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeRender.o -c volumeRender.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeRender_kernel.o -c volumeRender_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o volumeRender volumeRender.o volumeRender_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp volumeRender ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/volumeRender'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/marchingCubes'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o marchingCubes.o -c marchingCubes.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o marchingCubes_kernel.o -c marchingCubes_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o marchingCubes marchingCubes.o marchingCubes_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp marchingCubes ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/marchingCubes'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGLES_EGLOutput'
    >>> WARNING - simpleGLES_EGLOutput is not supported on Linux x86_64 - waiving sample <<<
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    >>> WARNING - gl31.h not found, please install gl31.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DUSE_CUDAINTEROP -DGRAPHICS_SETUP_EGL -DUSE_GLES -I/usr/include/libdrm -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGLES_EGLOutput.o -c simpleGLES_EGLOutput.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGLES_EGLOutput simpleGLES_EGLOutput.o -lGLESv2 -lEGL -ldrm
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleGLES_EGLOutput ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGLES_EGLOutput'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleTexture3D'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTexture3D.o -c simpleTexture3D.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTexture3D_kernel.o -c simpleTexture3D_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleTexture3D simpleTexture3D.o simpleTexture3D_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleTexture3D ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleTexture3D'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGLES_screen'
    >>> WARNING - simpleGLES_screen is not supported on Linux x86_64 - waiving sample <<<
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    >>> WARNING - gl31.h not found, please install gl31.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DUSE_CUDAINTEROP -DGRAPHICS_SETUP_EGL -DUSE_GLES -DWIN_INTERFACE_CUSTOM -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGLES_screen.o -c simpleGLES_screen.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGLES_screen simpleGLES_screen.o -lGLESv2 -lEGL -lscreen
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleGLES_screen ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGLES_screen'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGLES'
    >>> WARNING - simpleGLES is not supported on Linux x86_64 - waiving sample <<<
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    >>> WARNING - gl31.h not found, please install gl31.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DUSE_CUDAINTEROP -DGRAPHICS_SETUP_EGL -DUSE_GLES -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGLES.o -c simpleGLES.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleGLES simpleGLES.o -lGLESv2 -lEGL -lX11
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleGLES ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/2_Graphics/simpleGLES'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/EGLStreams_CUDA_Interop'
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuda_consumer.o -c cuda_consumer.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuda_producer.o -c cuda_producer.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o eglstrm_common.o -c eglstrm_common.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o EGLStreams_CUDA_Interop cuda_consumer.o cuda_producer.o eglstrm_common.o main.o -lEGL -lcuda
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp EGLStreams_CUDA_Interop ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/EGLStreams_CUDA_Interop'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/dxtc'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dxtc.o -c dxtc.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dxtc dxtc.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp dxtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/dxtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/boxFilter'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o boxFilter.o -c boxFilter.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o boxFilter_cpu.o -c boxFilter_cpu.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o boxFilter_kernel.o -c boxFilter_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o boxFilter boxFilter.o boxFilter_cpu.o boxFilter_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp boxFilter ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/boxFilter'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/SobelFilter'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o SobelFilter.o -c SobelFilter.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o SobelFilter_kernels.o -c SobelFilter_kernels.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o SobelFilter SobelFilter.o SobelFilter_kernels.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp SobelFilter ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/SobelFilter'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/convolutionSeparable'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionSeparable.o -c convolutionSeparable.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionSeparable_gold.o -c convolutionSeparable_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionSeparable convolutionSeparable.o convolutionSeparable_gold.o main.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp convolutionSeparable ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/convolutionSeparable'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/imageDenoising'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bmploader.o -c bmploader.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o imageDenoising.o -c imageDenoising.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o imageDenoisingGL.o -c imageDenoisingGL.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o imageDenoising bmploader.o imageDenoising.o imageDenoisingGL.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp imageDenoising ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/imageDenoising'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/dct8x8'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o BmpUtil.o -c BmpUtil.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o DCT8x8_Gold.o -c DCT8x8_Gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dct8x8.o -c dct8x8.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dct8x8 BmpUtil.o DCT8x8_Gold.o dct8x8.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp dct8x8 ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/dct8x8'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/recursiveGaussian'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o recursiveGaussian.o -c recursiveGaussian.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o recursiveGaussian_cuda.o -c recursiveGaussian_cuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o recursiveGaussian recursiveGaussian.o recursiveGaussian_cuda.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp recursiveGaussian ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/recursiveGaussian'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/bilateralFilter'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bilateralFilter.o -c bilateralFilter.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bilateralFilter_cpu.o -c bilateralFilter_cpu.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bilateral_kernel.o -c bilateral_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bmploader.o -c bmploader.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bilateralFilter bilateralFilter.o bilateralFilter_cpu.o bilateral_kernel.o bmploader.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp bilateralFilter ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/bilateralFilter'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/dwtHaar1D'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dwtHaar1D.o -c dwtHaar1D.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dwtHaar1D dwtHaar1D.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp dwtHaar1D ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/dwtHaar1D'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/postProcessGL'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o postProcessGL.o -c postProcessGL.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o postProcessGL main.o postProcessGL.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp postProcessGL ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/postProcessGL'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/convolutionTexture'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionTexture.o -c convolutionTexture.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionTexture_gold.o -c convolutionTexture_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionTexture convolutionTexture.o convolutionTexture_gold.o main.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp convolutionTexture ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/convolutionTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/stereoDisparity'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o stereoDisparity.o -c stereoDisparity.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o stereoDisparity stereoDisparity.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp stereoDisparity ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/stereoDisparity'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/EGLStream_CUDA_CrossGPU'
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuda_consumer.c.o -c cuda_consumer.c
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuda_producer.c.o -c cuda_producer.c
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o eglstrm_common.c.o -c eglstrm_common.c
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o kernel.o -c kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.c.o -c main.c
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o EGLStream_CUDA_CrossGPU cuda_consumer.c.o cuda_producer.c.o eglstrm_common.c.o kernel.o main.c.o -lEGL -lcuda
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp EGLStream_CUDA_CrossGPU ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/EGLStream_CUDA_CrossGPU'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/simpleCUDA2GL'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUDA2GL.o -c simpleCUDA2GL.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUDA2GL main.o simpleCUDA2GL.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp simpleCUDA2GL ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/simpleCUDA2GL'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/histogram'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o histogram256.o -c histogram256.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o histogram64.o -c histogram64.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o histogram_gold.o -c histogram_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o histogram histogram256.o histogram64.o histogram_gold.o main.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp histogram ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/histogram'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/cudaDecodeGL'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o FrameQueue.o -c FrameQueue.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o ImageGL.o -c ImageGL.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o VideoDecoder.o -c VideoDecoder.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o VideoParser.o -c VideoParser.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o VideoSource.o -c VideoSource.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o cudaModuleMgr.o -c cudaModuleMgr.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o cudaProcessFrame.o -c cudaProcessFrame.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o dynlink_cuda.o -c dynlink_cuda.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o dynlink_nvcuvid.o -c dynlink_nvcuvid.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o videoDecodeGL.o -c videoDecodeGL.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o cudaDecodeGL FrameQueue.o ImageGL.o VideoDecoder.o VideoParser.o VideoSource.o cudaModuleMgr.o cudaProcessFrame.o dynlink_cuda.o dynlink_nvcuvid.o videoDecodeGL.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut -lcuda -L../../common/lib/linux/x86_64 -lcudart -lnvcuvid -lGLEW
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp cudaDecodeGL ../../bin/x86_64/linux/release
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -DINIT_CUDA_GL -Xcompiler -no-pie -gencode arch=compute_30,code=compute_30 -o NV12ToARGB_drvapi64.ptx -ptx NV12ToARGB_drvapi.cu
    [@] mkdir -p data
    [@] cp -f NV12ToARGB_drvapi64.ptx ./data
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp -f NV12ToARGB_drvapi64.ptx ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/cudaDecodeGL'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/HSOpticalFlow'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o flowCUDA.o -c flowCUDA.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o flowGold.o -c flowGold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o HSOpticalFlow flowCUDA.o flowGold.o main.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp HSOpticalFlow ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/HSOpticalFlow'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/convolutionFFT2D'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionFFT2D.o -c convolutionFFT2D.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionFFT2D_gold.o -c convolutionFFT2D_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o convolutionFFT2D convolutionFFT2D.o convolutionFFT2D_gold.o main.o  -lcufft
    mkdir -p ../../bin/x86_64/linux/release
    cp convolutionFFT2D ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/convolutionFFT2D'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/3_Imaging/bicubicTexture'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bicubicTexture.o -c bicubicTexture.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bicubicTexture_cuda.o -c bicubicTexture_cuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bicubicTexture bicubicTexture.o bicubicTexture_cuda.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp bicubicTexture ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/3_Imaging/bicubicTexture'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/quasirandomGenerator_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o quasirandomGenerator.o -c quasirandomGenerator.cpp
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o quasirandomGenerator_gold.o -c quasirandomGenerator_gold.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o quasirandomGenerator_nvrtc quasirandomGenerator.o quasirandomGenerator_gold.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp quasirandomGenerator_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/quasirandomGenerator_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/quasirandomGenerator'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o quasirandomGenerator.o -c quasirandomGenerator.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o quasirandomGenerator_gold.o -c quasirandomGenerator_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o quasirandomGenerator_kernel.o -c quasirandomGenerator_kernel.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o quasirandomGenerator quasirandomGenerator.o quasirandomGenerator_gold.o quasirandomGenerator_kernel.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp quasirandomGenerator ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/quasirandomGenerator'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/binomialOptions'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o binomialOptions.o -c binomialOptions.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o binomialOptions_gold.o -c binomialOptions_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o binomialOptions_kernel.o -c binomialOptions_kernel.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o binomialOptions binomialOptions.o binomialOptions_gold.o binomialOptions_kernel.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp binomialOptions ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/binomialOptions'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/BlackScholes_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o BlackScholes.o -c BlackScholes.cpp
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o BlackScholes_gold.o -c BlackScholes_gold.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o BlackScholes_nvrtc BlackScholes.o BlackScholes_gold.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp BlackScholes_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/BlackScholes_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/SobolQRNG'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o sobol.o -c sobol.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o sobol_gold.o -c sobol_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o sobol_gpu.o -c sobol_gpu.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o sobol_primitives.o -c sobol_primitives.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o SobolQRNG sobol.o sobol_gold.o sobol_gpu.o sobol_primitives.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp SobolQRNG ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/SobolQRNG'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/BlackScholes'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -maxrregcount=16 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o BlackScholes.o -c BlackScholes.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -maxrregcount=16 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o BlackScholes_gold.o -c BlackScholes_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o BlackScholes BlackScholes.o BlackScholes_gold.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp BlackScholes ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/BlackScholes'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/MonteCarloMultiGPU'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MonteCarloMultiGPU.o -c MonteCarloMultiGPU.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MonteCarlo_gold.o -c MonteCarlo_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MonteCarlo_kernel.o -c MonteCarlo_kernel.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o multithreading.o -c multithreading.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MonteCarloMultiGPU MonteCarloMultiGPU.o MonteCarlo_gold.o MonteCarlo_kernel.o multithreading.o  -lcurand
    mkdir -p ../../bin/x86_64/linux/release
    cp MonteCarloMultiGPU ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/MonteCarloMultiGPU'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/4_Finance/binomialOptions_nvrtc'
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o binomialOptions.o -c binomialOptions.cpp
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o binomialOptions_gold.o -c binomialOptions_gold.cpp
    g++ -I../../common/inc -I"/usr/local/cuda"/include   -o binomialOptions_gpu.o -c binomialOptions_gpu.cpp
    g++  -L"/usr/local/cuda"/lib64/stubs -L"/usr/local/cuda"/lib64 -o binomialOptions_nvrtc binomialOptions.o binomialOptions_gold.o binomialOptions_gpu.o  -lcuda -lnvrtc
    mkdir -p ../../bin/x86_64/linux/release
    cp binomialOptions_nvrtc ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/4_Finance/binomialOptions_nvrtc'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/nbody'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bodysystemcuda.o -c bodysystemcuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nbody.o -c nbody.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o render_particles.o -c render_particles.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nbody bodysystemcuda.o nbody.o render_particles.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp nbody ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/nbody'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/smokeParticles'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o GLSLProgram.o -c GLSLProgram.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o ParticleSystem.o -c ParticleSystem.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o ParticleSystem_cuda.o -c ParticleSystem_cuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o SmokeRenderer.o -c SmokeRenderer.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o SmokeShaders.o -c SmokeShaders.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o framebufferObject.o -c framebufferObject.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o particleDemo.o -c particleDemo.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o renderbuffer.o -c renderbuffer.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o smokeParticles GLSLProgram.o ParticleSystem.o ParticleSystem_cuda.o SmokeRenderer.o SmokeShaders.o framebufferObject.o particleDemo.o renderbuffer.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp smokeParticles ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/smokeParticles'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/fluidsGL'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fluidsGL.o -c fluidsGL.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fluidsGL_kernels.o -c fluidsGL_kernels.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fluidsGL fluidsGL.o fluidsGL_kernels.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut -lcufft
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp fluidsGL ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/fluidsGL'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/particles'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o particleSystem.o -c particleSystem.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o particleSystem_cuda.o -c particleSystem_cuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o particles.o -c particles.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o render_particles.o -c render_particles.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o shaders.o -c shaders.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o particles particleSystem.o particleSystem_cuda.o particles.o render_particles.o shaders.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp particles ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/particles'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/oceanFFT'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o oceanFFT.o -c oceanFFT.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o oceanFFT_kernel.o -c oceanFFT_kernel.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o oceanFFT oceanFFT.o oceanFFT_kernel.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut -lcufft
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp oceanFFT ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/oceanFFT'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/nbody_opengles'
    >>> WARNING - nbody_opengles is not supported on Linux x86_64 - waiving sample <<<
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    >>> WARNING - gl31.h not found, please install gl31.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bodysystemcuda.o -c bodysystemcuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nbody_opengles.o -c nbody_opengles.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o render_particles.o -c render_particles.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nbody_opengles bodysystemcuda.o nbody_opengles.o render_particles.o -lGLESv2 -lEGL -lX11
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp nbody_opengles ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/nbody_opengles'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/nbody_screen'
    >>> WARNING - nbody_screen is not supported on Linux x86_64 - waiving sample <<<
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    >>> WARNING - gl31.h not found, please install gl31.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bodysystemcuda.o -c bodysystemcuda.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nbody_screen.o -c nbody_screen.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -ftz=true -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o render_particles.o -c render_particles.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nbody_screen bodysystemcuda.o nbody_screen.o render_particles.o -lGLESv2 -lEGL -lscreen
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp nbody_screen ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/nbody_screen'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/5_Simulations/fluidsGLES'
    >>> WARNING - fluidsGLES is not supported on Linux x86_64 - waiving sample <<<
    >>> WARNING - egl.h not found, please install egl.h <<<
    >>> WARNING - eglext.h not found, please install eglext.h <<<
    >>> WARNING - gl31.h not found, please install gl31.h <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fluidsGLES.o -c fluidsGLES.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fluidsGLES_kernels.o -c fluidsGLES_kernels.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fluidsGLES fluidsGLES.o fluidsGLES_kernels.o -lGLESv2 -lEGL -lX11 -lcufft
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp fluidsGLES ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/5_Simulations/fluidsGLES'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/newdelete'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o newdelete.o -c newdelete.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o newdelete newdelete.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp newdelete ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/newdelete'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/radixSortThrust'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o radixSortThrust.o -c radixSortThrust.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o radixSortThrust radixSortThrust.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp radixSortThrust ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/radixSortThrust'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/eigenvalues'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bisect_large.o -c bisect_large.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bisect_small.o -c bisect_small.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bisect_util.o -c bisect_util.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o gerschgorin.o -c gerschgorin.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o matlab.o -c matlab.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o eigenvalues bisect_large.o bisect_small.o bisect_util.o gerschgorin.o main.o matlab.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp eigenvalues ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/eigenvalues'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpAdvancedQuicksort'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpAdvancedQuicksort.o -c cdpAdvancedQuicksort.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpBitonicSort.o -c cdpBitonicSort.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpAdvancedQuicksort cdpAdvancedQuicksort.o cdpBitonicSort.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp cdpAdvancedQuicksort ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpAdvancedQuicksort'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpLUDecomposition'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdp_lu.o -c cdp_lu.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdp_lu_main.o -c cdp_lu_main.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dgetf2.o -c dgetf2.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dgetrf.o -c dgetrf.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o dlaswp.o -c dlaswp.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpLUDecomposition cdp_lu.o cdp_lu_main.o dgetf2.o dgetrf.o dlaswp.o  -lcublas -lcublas_device -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp cdpLUDecomposition ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpLUDecomposition'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/alignedTypes'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o alignedTypes.o -c alignedTypes.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o alignedTypes alignedTypes.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp alignedTypes ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/alignedTypes'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/FDTD3d'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -Iinc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FDTD3d.o -c src/FDTD3d.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -Iinc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FDTD3dGPU.o -c src/FDTD3dGPU.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -Iinc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FDTD3dReference.o -c src/FDTD3dReference.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FDTD3d FDTD3d.o FDTD3dGPU.o FDTD3dReference.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp FDTD3d ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/FDTD3d'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/segmentationTreeThrust'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o segmentationTree.o -c segmentationTree.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o segmentationTreeThrust segmentationTree.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp segmentationTreeThrust ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/segmentationTreeThrust'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/scalarProd'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o scalarProd.o -c scalarProd.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o scalarProd_cpu.o -c scalarProd_cpu.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o scalarProd scalarProd.o scalarProd_cpu.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp scalarProd ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/scalarProd'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/transpose'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o transpose.o -c transpose.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o transpose transpose.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp transpose ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/transpose'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/simpleHyperQ'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleHyperQ.o -c simpleHyperQ.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleHyperQ simpleHyperQ.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleHyperQ ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/simpleHyperQ'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/FunctionPointers'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FunctionPointers.o -c FunctionPointers.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FunctionPointers_kernels.o -c FunctionPointers_kernels.cu
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o FunctionPointers FunctionPointers.o FunctionPointers_kernels.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp FunctionPointers ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/FunctionPointers'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpQuadtree'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpQuadtree.o -c cdpQuadtree.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpQuadtree cdpQuadtree.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp cdpQuadtree ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpQuadtree'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/scan'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o scan.o -c scan.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o scan_gold.o -c scan_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o scan main.o scan.o scan_gold.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp scan ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/scan'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/concurrentKernels'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o concurrentKernels.o -c concurrentKernels.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o concurrentKernels concurrentKernels.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp concurrentKernels ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/concurrentKernels'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/shfl_scan'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -O3 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o shfl_scan.o -c shfl_scan.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o shfl_scan shfl_scan.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp shfl_scan ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/shfl_scan'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/lineOfSight'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o lineOfSight.o -c lineOfSight.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o lineOfSight lineOfSight.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp lineOfSight ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/lineOfSight'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/sortingNetworks'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bitonicSort.o -c bitonicSort.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o oddEvenMergeSort.o -c oddEvenMergeSort.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o sortingNetworks_validate.o -c sortingNetworks_validate.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o sortingNetworks bitonicSort.o main.o oddEvenMergeSort.o sortingNetworks_validate.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp sortingNetworks ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/sortingNetworks'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/fastWalshTransform'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fastWalshTransform.o -c fastWalshTransform.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fastWalshTransform_gold.o -c fastWalshTransform_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o fastWalshTransform fastWalshTransform.o fastWalshTransform_gold.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp fastWalshTransform ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/fastWalshTransform'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/matrixMulDynlinkJIT'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o cuda_drvapi_dynlink.c.o -c cuda_drvapi_dynlink.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o matrixMulDynlinkJIT.o -c matrixMulDynlinkJIT.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o matrixMul_gold.o -c matrixMul_gold.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o matrixMul_kernel_32_ptxdump.c.o -c matrixMul_kernel_32_ptxdump.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o matrixMul_kernel_64_ptxdump.c.o -c matrixMul_kernel_64_ptxdump.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o matrixMulDynlinkJIT cuda_drvapi_dynlink.c.o matrixMulDynlinkJIT.o matrixMul_gold.o matrixMul_kernel_32_ptxdump.c.o matrixMul_kernel_64_ptxdump.c.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp matrixMulDynlinkJIT ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/matrixMulDynlinkJIT'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/interval'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I.  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o interval.o -c interval.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o interval interval.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp interval ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/interval'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/c++11_cuda'
    >>> GCC Version is greater or equal to 4.7.0 <<<
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    --std=c++11 -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o c++11_cuda.o -c c++11_cuda.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o c++11_cuda c++11_cuda.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp c++11_cuda ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/c++11_cuda'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/reductionMultiBlockCG'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o reductionMultiBlockCG.o -c reductionMultiBlockCG.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o reductionMultiBlockCG reductionMultiBlockCG.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp reductionMultiBlockCG ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/reductionMultiBlockCG'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpBezierTessellation'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o BezierLineCDP.o -c BezierLineCDP.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cdpBezierTessellation BezierLineCDP.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp cdpBezierTessellation ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/cdpBezierTessellation'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/StreamPriorities'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o StreamPriorities.o -c StreamPriorities.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o StreamPriorities StreamPriorities.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp StreamPriorities ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/StreamPriorities'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/conjugateGradientMultiBlockCG'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o conjugateGradientMultiBlockCG.o -c conjugateGradientMultiBlockCG.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o conjugateGradientMultiBlockCG conjugateGradientMultiBlockCG.o  -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp conjugateGradientMultiBlockCG ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/conjugateGradientMultiBlockCG'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/threadMigration'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o threadMigration.o -c threadMigration.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o threadMigration threadMigration.o  -lcuda -lcudart_static
    mkdir -p ../../bin/x86_64/linux/release
    cp threadMigration ../../bin/x86_64/linux/release
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o threadMigration_kernel64.ptx -ptx threadMigration_kernel.cu
    mkdir -p data
    cp -f threadMigration_kernel64.ptx ./data
    mkdir -p ../../bin/x86_64/linux/release
    cp -f threadMigration_kernel64.ptx ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/threadMigration'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/threadFenceReduction'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o threadFenceReduction.o -c threadFenceReduction.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o threadFenceReduction threadFenceReduction.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp threadFenceReduction ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/threadFenceReduction'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/reduction'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o reduction.o -c reduction.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o reduction_kernel.o -c reduction_kernel.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o reduction reduction.o reduction_kernel.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp reduction ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/reduction'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/mergeSort'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o bitonic.o -c bitonic.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mergeSort.o -c mergeSort.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mergeSort_host.o -c mergeSort_host.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mergeSort_validate.o -c mergeSort_validate.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mergeSort bitonic.o main.o mergeSort.o mergeSort_host.o mergeSort_validate.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp mergeSort ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/mergeSort'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/ptxjit'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o ptxjit.o -c ptxjit.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o ptxjit ptxjit.o  -lcuda -lcudart_static
    mkdir -p ../../bin/x86_64/linux/release
    cp ptxjit ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/ptxjit'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/6_Advanced/warpAggregatedAtomicsCG'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o warpAggregatedAtomicsCG.o -c warpAggregatedAtomicsCG.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o warpAggregatedAtomicsCG warpAggregatedAtomicsCG.o 
    mkdir -p ../../bin/x86_64/linux/release
    cp warpAggregatedAtomicsCG ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/6_Advanced/warpAggregatedAtomicsCG'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiQ'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c src/main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o piestimator.o -c src/piestimator.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o test.o -c src/test.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MC_EstimatePiQ main.o piestimator.o test.o  -lcurand
    mkdir -p ../../bin/x86_64/linux/release
    cp MC_EstimatePiQ ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiQ'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_Pagerank'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_Pagerank.o -c nvgraph_Pagerank.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_Pagerank nvgraph_Pagerank.o  -lnvgraph
    mkdir -p ../../bin/x86_64/linux/release
    cp nvgraph_Pagerank ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_Pagerank'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_SpectralClustering'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_SpectralClustering.o -c nvgraph_SpectralClustering.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_SpectralClustering nvgraph_SpectralClustering.o  -lnvgraph
    mkdir -p ../../bin/x86_64/linux/release
    cp nvgraph_SpectralClustering ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_SpectralClustering'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cannyEdgeDetectorNPP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I../common/UtilNPP -I../common/FreeImage/include  -m64    -gencode arch=compute_30,code=compute_30 -o cannyEdgeDetectorNPP.o -c cannyEdgeDetectorNPP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o cannyEdgeDetectorNPP cannyEdgeDetectorNPP.o  -L../common/FreeImage/lib/ -L../common/FreeImage/lib/linux -L../common/FreeImage/lib/linux/x86_64 -lnppisu_static -lnppif_static -lnppc_static -lculibos -lfreeimage
    mkdir -p ../../bin/x86_64/linux/release
    cp cannyEdgeDetectorNPP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cannyEdgeDetectorNPP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_SingleAsianOptionP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c src/main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o pricingengine.o -c src/pricingengine.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o test.o -c src/test.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MC_SingleAsianOptionP main.o pricingengine.o test.o  -lcurand
    mkdir -p ../../bin/x86_64/linux/release
    cp MC_SingleAsianOptionP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_SingleAsianOptionP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/conjugateGradient'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o conjugateGradient main.o  -lcublas_static -lcusparse_static -lculibos
    mkdir -p ../../bin/x86_64/linux/release
    cp conjugateGradient ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/conjugateGradient'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverSp_LowlevelQR'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverSp_LowlevelQR.o -c cuSolverSp_LowlevelQR.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio.c.o -c mmio.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio_wrapper.o -c mmio_wrapper.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\"   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverSp_LowlevelQR cuSolverSp_LowlevelQR.o mmio.c.o mmio_wrapper.o  -lcusolver -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp cuSolverSp_LowlevelQR ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverSp_LowlevelQR'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/BiCGStab'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio.c.o -c mmio.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o pbicgstab.o -c pbicgstab.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o BiCGStab mmio.c.o pbicgstab.o  -lcusparse -lcublas
    mkdir -p ../../bin/x86_64/linux/release
    cp BiCGStab ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/BiCGStab'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT_MGPU'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT_MGPU.o -c simpleCUFFT_MGPU.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT_MGPU simpleCUFFT_MGPU.o  -lcufft
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCUFFT_MGPU ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT_MGPU'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_SemiRingSpMV'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_SemiRingSpMV.o -c nvgraph_SemiRingSpMV.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_SemiRingSpMV nvgraph_SemiRingSpMV.o  -lnvgraph
    mkdir -p ../../bin/x86_64/linux/release
    cp nvgraph_SemiRingSpMV ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_SemiRingSpMV'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuHook'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuHook.o -c cuHook.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuHook cuHook.o  -lcuda -L"/usr/local/cuda"/lib64 -ldl
    mkdir -p ../../bin/x86_64/linux/release
    cp cuHook ../../bin/x86_64/linux/release
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    --compiler-options '-fPIC' -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o libcuhook.o -c libcuhook.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -shared   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o libcuhook.so.1 libcuhook.o  -lcuda -L"/usr/local/cuda"/lib64 -ldl
    mkdir -p ../../bin/x86_64/linux/release
    cp libcuhook.so.1 ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuHook'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/conjugateGradientPrecond'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o conjugateGradientPrecond main.o  -lcublas -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp conjugateGradientPrecond ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/conjugateGradientPrecond'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/freeImageInteropNPP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I../common/UtilNPP -I../common/FreeImage/include  -m64    -gencode arch=compute_30,code=compute_30 -o freeImageInteropNPP.o -c freeImageInteropNPP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o freeImageInteropNPP freeImageInteropNPP.o  -L../common/FreeImage/lib -L../common/FreeImage/lib/linux -L../common/FreeImage/lib/linux/x86_64 -lnppisu -lnppif -lnppc -lfreeimage
    mkdir -p ../../bin/x86_64/linux/release
    cp freeImageInteropNPP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/freeImageInteropNPP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/batchCUBLAS'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o batchCUBLAS.o -c batchCUBLAS.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o batchCUBLAS batchCUBLAS.o  -lcublas
    mkdir -p ../../bin/x86_64/linux/release
    cp batchCUBLAS ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/batchCUBLAS'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiInlineQ'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c src/main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o piestimator.o -c src/piestimator.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o test.o -c src/test.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MC_EstimatePiInlineQ main.o piestimator.o test.o  -lcurand
    mkdir -p ../../bin/x86_64/linux/release
    cp MC_EstimatePiInlineQ ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiInlineQ'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiInlineP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c src/main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o piestimator.o -c src/piestimator.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o test.o -c src/test.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MC_EstimatePiInlineP main.o piestimator.o test.o  -lcurand
    mkdir -p ../../bin/x86_64/linux/release
    cp MC_EstimatePiInlineP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiInlineP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverSp_LowlevelCholesky'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverSp_LowlevelCholesky.o -c cuSolverSp_LowlevelCholesky.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio.c.o -c mmio.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio_wrapper.o -c mmio_wrapper.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\"   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverSp_LowlevelCholesky cuSolverSp_LowlevelCholesky.o mmio.c.o mmio_wrapper.o  -lcusolver -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp cuSolverSp_LowlevelCholesky ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverSp_LowlevelCholesky'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT_callback'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT_callback.o -c simpleCUFFT_callback.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT_callback simpleCUFFT_callback.o  -lcufft_static -lculibos
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCUFFT_callback ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT_callback'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/histEqualizationNPP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I../common/UtilNPP -I../common/FreeImage/include  -m64    -gencode arch=compute_30,code=compute_30 -o histEqualizationNPP.o -c histEqualizationNPP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o histEqualizationNPP histEqualizationNPP.o  -L../common/FreeImage/lib -L../common/FreeImage/lib/linux -L../common/FreeImage/lib/linux/x86_64 -lnppisu -lnppist -lnppicc -lnppc -lfreeimage
    mkdir -p ../../bin/x86_64/linux/release
    cp histEqualizationNPP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/histEqualizationNPP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/boxFilterNPP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I../common/UtilNPP -I../common/FreeImage/include  -m64    -gencode arch=compute_30,code=compute_30 -o boxFilterNPP.o -c boxFilterNPP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o boxFilterNPP boxFilterNPP.o  -L../common/FreeImage/lib/ -L../common/FreeImage/lib/linux -L../common/FreeImage/lib/linux/x86_64 -lnppisu_static -lnppif_static -lnppc_static -lculibos -lfreeimage
    mkdir -p ../../bin/x86_64/linux/release
    cp boxFilterNPP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/boxFilterNPP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverSp_LinearSolver'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverSp_LinearSolver.o -c cuSolverSp_LinearSolver.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio.c.o -c mmio.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio_wrapper.o -c mmio_wrapper.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\"   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverSp_LinearSolver cuSolverSp_LinearSolver.o mmio.c.o mmio_wrapper.o  -lcusolver -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp cuSolverSp_LinearSolver ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverSp_LinearSolver'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_SSSP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_SSSP.o -c nvgraph_SSSP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o nvgraph_SSSP nvgraph_SSSP.o  -lnvgraph
    mkdir -p ../../bin/x86_64/linux/release
    cp nvgraph_SSSP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/nvgraph_SSSP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c src/main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o piestimator.o -c src/piestimator.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o test.o -c src/test.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o MC_EstimatePiP main.o piestimator.o test.o  -lcurand
    mkdir -p ../../bin/x86_64/linux/release
    cp MC_EstimatePiP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MC_EstimatePiP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT.o -c simpleCUFFT.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT simpleCUFFT.o  -lcufft
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCUFFT ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverDn_LinearSolver'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverDn_LinearSolver.o -c cuSolverDn_LinearSolver.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio.c.o -c mmio.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio_wrapper.o -c mmio_wrapper.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\"   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverDn_LinearSolver cuSolverDn_LinearSolver.o mmio.c.o mmio_wrapper.o  -lcusolver -lcublas -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp cuSolverDn_LinearSolver ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverDn_LinearSolver'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverRf'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverRf.o -c cuSolverRf.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio.c.o -c mmio.c
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\" -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o mmio_wrapper.o -c mmio_wrapper.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64    -Xcompiler -fopenmp -Xcompiler \"-Wl,--no-as-needed\"   -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o cuSolverRf cuSolverRf.o mmio.c.o mmio_wrapper.o  -lcusolver -lcublas -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp cuSolverRf ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/cuSolverRf'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/jpegNPP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I../common/UtilNPP -I../common/FreeImage/include  -m64    -gencode arch=compute_30,code=compute_30 -o jpegNPP.o -c jpegNPP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o jpegNPP jpegNPP.o  -L../common/FreeImage/lib -L../common/FreeImage/lib/linux -L../common/FreeImage/lib/linux/x86_64 -lnppisu -lnppicom -lnppig -lnppc -lfreeimage
    mkdir -p ../../bin/x86_64/linux/release
    cp jpegNPP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/jpegNPP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleDevLibCUBLAS'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o kernels.o -c kernels.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -dc -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleDevLibCUBLAS.o -c simpleDevLibCUBLAS.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleDevLibCUBLAS kernels.o simpleDevLibCUBLAS.o  -lcublas -lcublas_device -lcudadevrt
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleDevLibCUBLAS ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleDevLibCUBLAS'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/FilterBorderControlNPP'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc -I../common/UtilNPP -I../common/FreeImage/include  -m64    -gencode arch=compute_30,code=compute_30 -o FilterBorderControlNPP.o -c FilterBorderControlNPP.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o FilterBorderControlNPP FilterBorderControlNPP.o  -L../common/FreeImage/lib/ -L../common/FreeImage/lib/linux -L../common/FreeImage/lib/linux/x86_64 -lnppisu_static -lnppif_static -lnppitc_static -lnppidei_static -lnppial_static -lnppc_static -lculibos -lfreeimage
    mkdir -p ../../bin/x86_64/linux/release
    cp FilterBorderControlNPP ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/FilterBorderControlNPP'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUBLAS'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o simpleCUBLAS.o -c simpleCUBLAS.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o simpleCUBLAS simpleCUBLAS.o  -lcublas
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCUBLAS ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUBLAS'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/randomFog'
    >>> WARNING - libGLU.so not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    >>> WARNING - glu.h not found, refer to CUDA Getting Started Guide for how to find and install them. <<<
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -Xcompiler -fpermissive -gencode arch=compute_30,code=compute_30 -o randomFog.o -c randomFog.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -I../../common/inc -m64 -Xcompiler -fpermissive -gencode arch=compute_30,code=compute_30 -o rng.o -c rng.cpp
    [@] /usr/local/cuda/bin/nvcc -ccbin g++ -m64 -gencode arch=compute_30,code=compute_30 -o randomFog randomFog.o rng.o -L/usr/lib/nvidia-compute-utils-410 -lGL -lGLU -lX11 -lglut -lcurand
    [@] mkdir -p ../../bin/x86_64/linux/release
    [@] cp randomFog ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/randomFog'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUBLASXT'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o simpleCUBLASXT.o -c simpleCUBLASXT.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o simpleCUBLASXT simpleCUBLASXT.o  -lcublas
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCUBLASXT ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUBLASXT'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT_2d_MGPU'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT_2d_MGPU.o -c simpleCUFFT_2d_MGPU.cu
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o simpleCUFFT_2d_MGPU simpleCUFFT_2d_MGPU.o  -lcufft
    mkdir -p ../../bin/x86_64/linux/release
    cp simpleCUFFT_2d_MGPU ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/simpleCUFFT_2d_MGPU'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/conjugateGradientUM'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o main.o -c main.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_37,code=sm_37 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_52,code=sm_52 -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_70,code=compute_70 -o conjugateGradientUM main.o  -lcublas -lcusparse
    mkdir -p ../../bin/x86_64/linux/release
    cp conjugateGradientUM ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/conjugateGradientUM'
    make[1]: Entering directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MersenneTwisterGP11213'
    "/usr/local/cuda"/bin/nvcc -ccbin g++ -I../../common/inc  -m64    -gencode arch=compute_30,code=compute_30 -o MersenneTwister.o -c MersenneTwister.cpp
    "/usr/local/cuda"/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=compute_30 -o MersenneTwisterGP11213 MersenneTwister.o  -lcurand_static -lculibos
    mkdir -p ../../bin/x86_64/linux/release
    cp MersenneTwisterGP11213 ../../bin/x86_64/linux/release
    make[1]: Leaving directory '/usr/local/cuda-9.0/samples/7_CUDALibraries/MersenneTwisterGP11213'
    Finished building CUDA samples


    In file included from simpleAssert.cpp:26:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    In file included from matrixMul.cpp:36:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    In file included from simpleVoteIntrinsics.cpp:18:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    In file included from clock.cpp:27:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    simpleSeparateCompilation.cu(90): warning: variable "devID" was set but never used
    
    In file included from vectorAdd.cpp:30:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is not valid on compute_70 and above, and should be replaced with __all_sync().To continue using __all(), specify virtual architecture compute_60 when targeting sm_70 and above, for example, using the pair of compiler options: -arch=compute_60 -code=sm_70.")
    
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is deprecated in favor of __all_sync() and may be removed in a future release (Use -Wno-deprecated-declarations to suppress this warning).")
    
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is deprecated in favor of __all_sync() and may be removed in a future release (Use -Wno-deprecated-declarations to suppress this warning).")
    
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is deprecated in favor of __all_sync() and may be removed in a future release (Use -Wno-deprecated-declarations to suppress this warning).")
    
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is deprecated in favor of __all_sync() and may be removed in a future release (Use -Wno-deprecated-declarations to suppress this warning).")
    
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is deprecated in favor of __all_sync() and may be removed in a future release (Use -Wno-deprecated-declarations to suppress this warning).")
    
    simpleVote_kernel.cuh(58): warning: function "all"
    /usr/local/cuda/bin/..//include/device_atomic_functions.hpp(216): here was declared deprecated ("__all() is deprecated in favor of __all_sync() and may be removed in a future release (Use -Wno-deprecated-declarations to suppress this warning).")
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    simpleAssert.cu(71): warning: variable "devID" was set but never used
    
    In file included from simpleAtomicIntrinsics.cpp:30:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    stereoDisparity.cu(33): warning: variable "sSDKsample" was declared but never referenced
    
    In file included from quasirandomGenerator_gpu.cuh:15:0,
                     from quasirandomGenerator.cpp:19:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    In file included from BlackScholes.cpp:19:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    ptxas warning : For profile sm_70 adjusting per thread register count of 16 to lower bound of 24
    In file included from binomialOptions_gpu.cpp:21:0:
    ../../common/inc/nvrtc_helper.h: In function ‘void compileFileToPTX(char*, int, char**, char**, size_t*, int)’:
    ../../common/inc/nvrtc_helper.h:50:29: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
             char *HeaderNames = "cooperative_groups.h";
                                 ^~~~~~~~~~~~~~~~~~~~~~
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    newdelete.cu(296): warning: variable "cuda_device" was set but never used
    
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    ptxas warning : Stack size for entry function '_Z20test_interval_newtonIdEvP12interval_gpuIT_EPiS2_i' cannot be statically determined
    pbicgstab.cpp: In function ‘int main(int, char**)’:
    pbicgstab.cpp:594:80: warning: ISO C++ forbids converting a string constant to ‘char*’ [-Wwrite-strings]
                                                DBICGSTAB_MAX_ULP_ERR, DBICGSTAB_EPS);
                                                                                    ^
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_35)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_37)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_50)
    @W@nvlink warning : Stack size for entry function '_Z23invokeDeviceCublasSgemmP14cublasStatus_tiPKfS2_S2_S2_Pf' cannot be statically determined (target: sm_50)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_52)
    @W@nvlink warning : Stack size for entry function '_Z23invokeDeviceCublasSgemmP14cublasStatus_tiPKfS2_S2_S2_Pf' cannot be statically determined (target: sm_52)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_60)
    @W@nvlink warning : Stack size for entry function '_Z23invokeDeviceCublasSgemmP14cublasStatus_tiPKfS2_S2_S2_Pf' cannot be statically determined (target: sm_60)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @O@ptxas info    : 'device-function-maxrregcount' is a BETA feature (target: sm_70)
    @W@nvlink warning : Stack size for entry function '_Z23invokeDeviceCublasSgemmP14cublasStatus_tiPKfS2_S2_S2_Pf' cannot be statically determined (target: sm_70)


If run into an error involving the GCC version:

```
$ sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-6 10
```

```
$ sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-6 10
```

And run again, otherwise, skip this step.

* After complete, **run a devicequery and bandwidth test**:


```bash
%%bash

cd /usr/local/cuda/samples/bin/x86_64/linux/release
./deviceQuery
```

    ./deviceQuery Starting...
    
     CUDA Device Query (Runtime API) version (CUDART static linking)
    
    Detected 1 CUDA Capable device(s)
    
    Device 0: "Tesla V100-SXM2-16GB"
      CUDA Driver Version / Runtime Version          10.0 / 9.0
      CUDA Capability Major/Minor version number:    7.0
      Total amount of global memory:                 16130 MBytes (16914055168 bytes)
      (80) Multiprocessors, ( 64) CUDA Cores/MP:     5120 CUDA Cores
      GPU Max Clock rate:                            1530 MHz (1.53 GHz)
      Memory Clock rate:                             877 Mhz
      Memory Bus Width:                              4096-bit
      L2 Cache Size:                                 6291456 bytes
      Maximum Texture Dimension Size (x,y,z)         1D=(131072), 2D=(131072, 65536), 3D=(16384, 16384, 16384)
      Maximum Layered 1D Texture Size, (num) layers  1D=(32768), 2048 layers
      Maximum Layered 2D Texture Size, (num) layers  2D=(32768, 32768), 2048 layers
      Total amount of constant memory:               65536 bytes
      Total amount of shared memory per block:       49152 bytes
      Total number of registers available per block: 65536
      Warp size:                                     32
      Maximum number of threads per multiprocessor:  2048
      Maximum number of threads per block:           1024
      Max dimension size of a thread block (x,y,z): (1024, 1024, 64)
      Max dimension size of a grid size    (x,y,z): (2147483647, 65535, 65535)
      Maximum memory pitch:                          2147483647 bytes
      Texture alignment:                             512 bytes
      Concurrent copy and kernel execution:          Yes with 2 copy engine(s)
      Run time limit on kernels:                     No
      Integrated GPU sharing Host Memory:            No
      Support host page-locked memory mapping:       Yes
      Alignment requirement for Surfaces:            Yes
      Device has ECC support:                        Enabled
      Device supports Unified Addressing (UVA):      Yes
      Supports Cooperative Kernel Launch:            Yes
      Supports MultiDevice Co-op Kernel Launch:      Yes
      Device PCI Domain ID / Bus ID / location ID:   0 / 0 / 4
      Compute Mode:
         < Default (multiple host threads can use ::cudaSetDevice() with device simultaneously) >
    
    deviceQuery, CUDA Driver = CUDART, CUDA Driver Version = 10.0, CUDA Runtime Version = 9.0, NumDevs = 1
    Result = PASS


## Install cuDNN 7.0

* **Download all 3 .deb files for CUDA9.0 and Ubuntu**

You will have to create a Nvidia account for this and go to the archive section of the cuDNN downloads

Ensure you download all 3 files:
- Runtime
- Developer
- Code Samples


**Unpackage the three files in this order**


```bash
%%bash 
sudo dpkg -i ~/libcudnn7_7.0.5.15-1+cuda9.0_amd64.deb
sudo dpkg -i ~/libcudnn7-dev_7.0.5.15-1+cuda9.0_amd64.deb
sudo dpkg -i ~/libcudnn7-doc_7.0.5.15-1+cuda9.0_amd64.deb
```

    (Reading database ... 94650 files and directories currently installed.)
    Preparing to unpack .../libcudnn7_7.0.5.15-1+cuda9.0_amd64.deb ...
    Unpacking libcudnn7 (7.0.5.15-1+cuda9.0) over (7.6.0.64-1+cuda9.0) ...
    Setting up libcudnn7 (7.0.5.15-1+cuda9.0) ...
    Processing triggers for libc-bin (2.27-3ubuntu1) ...
    (Reading database ... 94651 files and directories currently installed.)
    Preparing to unpack .../libcudnn7-dev_7.0.5.15-1+cuda9.0_amd64.deb ...
    update-alternatives: removing manually selected alternative - switching libcudnn to auto mode
    Unpacking libcudnn7-dev (7.0.5.15-1+cuda9.0) over (7.6.0.64-1+cuda9.0) ...
    Setting up libcudnn7-dev (7.0.5.15-1+cuda9.0) ...
    update-alternatives: using /usr/include/x86_64-linux-gnu/cudnn_v7.h to provide /usr/include/cudnn.h (libcudnn) in auto mode
    (Reading database ... 94651 files and directories currently installed.)
    Preparing to unpack .../libcudnn7-doc_7.0.5.15-1+cuda9.0_amd64.deb ...
    Unpacking libcudnn7-doc (7.0.5.15-1+cuda9.0) over (7.6.0.64-1+cuda9.0) ...
    Setting up libcudnn7-doc (7.0.5.15-1+cuda9.0) ...


    dpkg: warning: downgrading libcudnn7 from 7.6.0.64-1+cuda9.0 to 7.0.5.15-1+cuda9.0
    dpkg: warning: downgrading libcudnn7-dev from 7.6.0.64-1+cuda9.0 to 7.0.5.15-1+cuda9.0
    dpkg: warning: downgrading libcudnn7-doc from 7.6.0.64-1+cuda9.0 to 7.0.5.15-1+cuda9.0


* **Verify the install is successful with the MNIST example**

From the download folder. Copy the files to somewhere with write access: 


```python
! cp -r /usr/src/cudnn_samples_v7/ ~
```

**Go to the MNIST example code, compile and run it**


```bash
%%bash
cd ~/cudnn_samples_v7/mnistCUDNN
sudo make
sudo ./mnistCUDNN
```

    /usr/local/cuda/bin/nvcc -ccbin g++ -I/usr/local/cuda/include -IFreeImage/include  -m64    -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_53,code=sm_53 -gencode arch=compute_53,code=compute_53 -o fp16_dev.o -c fp16_dev.cu
    g++ -I/usr/local/cuda/include -IFreeImage/include   -o fp16_emu.o -c fp16_emu.cpp
    g++ -I/usr/local/cuda/include -IFreeImage/include   -o mnistCUDNN.o -c mnistCUDNN.cpp
    /usr/local/cuda/bin/nvcc -ccbin g++   -m64      -gencode arch=compute_30,code=sm_30 -gencode arch=compute_35,code=sm_35 -gencode arch=compute_50,code=sm_50 -gencode arch=compute_53,code=sm_53 -gencode arch=compute_53,code=compute_53 -o mnistCUDNN fp16_dev.o fp16_emu.o mnistCUDNN.o  -LFreeImage/lib/linux/x86_64 -LFreeImage/lib/linux -lcudart -lcublas -lcudnn -lfreeimage -lstdc++ -lm
    cudnnGetVersion() : 7005 , CUDNN_VERSION from cudnn.h : 7005 (7.0.5)
    Host compiler version : GCC 6.5.0
    There are 1 CUDA capable devices on your machine :
    device 0 : sms 80  Capabilities 7.0, SmClock 1530.0 Mhz, MemSize (Mb) 16130, MemClock 877.0 Mhz, Ecc=1, boardGroupID=0
    Using device 0
    
    Testing single precision
    Loading image data/one_28x28.pgm
    Performing forward propagation ...
    Testing cudnnGetConvolutionForwardAlgorithm ...
    Fastest algorithm is Algo 5
    Testing cudnnFindConvolutionForwardAlgorithm ...
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 0: 0.024576 time requiring 0 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 2: 0.053248 time requiring 57600 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 1: 0.056320 time requiring 3464 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 7: 0.078848 time requiring 2057744 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 5: 0.086016 time requiring 203008 memory
    Resulting weights from Softmax:
    0.0000000 0.9999399 0.0000000 0.0000000 0.0000561 0.0000000 0.0000012 0.0000017 0.0000010 0.0000000 
    Loading image data/three_28x28.pgm
    Performing forward propagation ...
    Resulting weights from Softmax:
    0.0000000 0.0000000 0.0000000 0.9999288 0.0000000 0.0000711 0.0000000 0.0000000 0.0000000 0.0000000 
    Loading image data/five_28x28.pgm
    Performing forward propagation ...
    Resulting weights from Softmax:
    0.0000000 0.0000008 0.0000000 0.0000002 0.0000000 0.9999820 0.0000154 0.0000000 0.0000012 0.0000006 
    
    Result of classification: 1 3 5
    
    Test passed!
    
    Testing half precision (math in single precision)
    Loading image data/one_28x28.pgm
    Performing forward propagation ...
    Testing cudnnGetConvolutionForwardAlgorithm ...
    Fastest algorithm is Algo 5
    Testing cudnnFindConvolutionForwardAlgorithm ...
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 0: 0.019456 time requiring 0 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 2: 0.049152 time requiring 28800 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 1: 0.051200 time requiring 3464 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 7: 0.067584 time requiring 2057744 memory
    ^^^^ CUDNN_STATUS_SUCCESS for Algo 5: 0.082944 time requiring 203008 memory
    Resulting weights from Softmax:
    0.0000001 1.0000000 0.0000001 0.0000000 0.0000563 0.0000001 0.0000012 0.0000017 0.0000010 0.0000001 
    Loading image data/three_28x28.pgm
    Performing forward propagation ...
    Resulting weights from Softmax:
    0.0000000 0.0000000 0.0000000 1.0000000 0.0000000 0.0000714 0.0000000 0.0000000 0.0000000 0.0000000 
    Loading image data/five_28x28.pgm
    Performing forward propagation ...
    Resulting weights from Softmax:
    0.0000000 0.0000008 0.0000000 0.0000002 0.0000000 1.0000000 0.0000154 0.0000000 0.0000012 0.0000006 
    
    Result of classification: 1 3 5
    
    Test passed!


# Configure CUDA and cuDNN

**Add LD_LIBRARY_PATH in your .bashrc file:**

Add the following line in the end or your .bashrc file export export:

```
LD_LIBRARY_PATH=/usr/local/cuda/lib64:${LD_LIBRARY_PATH:+:${LD_LIBRARY_PATH}}
```

And source it with source:

```
$ ~/.bashrc
```

# Install tensorflow with GPU

**Require v=1.12.0 as with CUDA 9.0**


```python
! pip3 install --upgrade tensorflow-gpu==1.12.0
```

    Requirement already up-to-date: tensorflow-gpu==1.12.0 in /usr/local/lib/python3.6/dist-packages
    Requirement already up-to-date: protobuf>=3.6.1 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: wheel>=0.26 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: numpy>=1.13.3 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: astor>=0.6.0 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: grpcio>=1.8.6 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: keras-preprocessing>=1.0.5 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: termcolor>=1.1.0 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: keras-applications>=1.0.6 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: six>=1.10.0 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: tensorboard<1.13.0,>=1.12.0 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: gast>=0.2.0 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: absl-py>=0.1.6 in /usr/local/lib/python3.6/dist-packages (from tensorflow-gpu==1.12.0)
    Requirement already up-to-date: setuptools in /usr/local/lib/python3.6/dist-packages (from protobuf>=3.6.1->tensorflow-gpu==1.12.0)
    Requirement already up-to-date: h5py in /usr/local/lib/python3.6/dist-packages (from keras-applications>=1.0.6->tensorflow-gpu==1.12.0)
    Requirement already up-to-date: werkzeug>=0.11.10 in /usr/local/lib/python3.6/dist-packages (from tensorboard<1.13.0,>=1.12.0->tensorflow-gpu==1.12.0)
    Requirement already up-to-date: markdown>=2.6.8 in /usr/local/lib/python3.6/dist-packages (from tensorboard<1.13.0,>=1.12.0->tensorflow-gpu==1.12.0)



```python
import tensorflow as tf
sess = tf.Session(config=tf.ConfigProto(log_device_placement=True))
```

* Wrap a Tensorflow MNIST python model for use as a prediction microservice in seldon-core
 
   * Run locally on Docker to test
   * Deploy on seldon-core running on minikube
 
## Dependencies

 * [Helm](https://github.com/kubernetes/helm)
 * [Minikube](https://github.com/kubernetes/minikube)
 * [S2I](https://github.com/openshift/source-to-image)

```bash
pip install seldon-core
```

## Train locally
 


```python
from tensorflow.examples.tutorials.mnist import input_data
mnist = input_data.read_data_sets("MNIST_data/", one_hot = True)
import tensorflow as tf

if __name__ == '__main__':
    
    x = tf.placeholder(tf.float32, [None,784], name="x")

    W = tf.Variable(tf.zeros([784,10]))
    b = tf.Variable(tf.zeros([10]))

    y = tf.nn.softmax(tf.matmul(x,W) + b, name="y")

    y_ = tf.placeholder(tf.float32, [None, 10])


    cross_entropy = tf.reduce_mean(-tf.reduce_sum(y_ * tf.log(y), reduction_indices=[1]))

    train_step = tf.train.GradientDescentOptimizer(0.5).minimize(cross_entropy)

    init = tf.initialize_all_variables()

    sess = tf.Session()
    sess.run(init)

    for i in range(1000):
        batch_xs, batch_ys = mnist.train.next_batch(100)
        sess.run(train_step, feed_dict={x: batch_xs, y_: batch_ys})

    correct_prediction = tf.equal(tf.argmax(y,1), tf.argmax(y_,1))
    accuracy = tf.reduce_mean(tf.cast(correct_prediction, tf.float32))
    print(sess.run(accuracy, feed_dict = {x: mnist.test.images, y_:mnist.test.labels}))

    saver = tf.train.Saver()

    saver.save(sess, "model/deep_mnist_model")


```

    WARNING:tensorflow:From <ipython-input-26-b7995d30f035>:2: read_data_sets (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use alternatives such as official/mnist/dataset.py from tensorflow/models.
    WARNING:tensorflow:From /usr/local/lib/python3.6/dist-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:260: maybe_download (from tensorflow.contrib.learn.python.learn.datasets.base) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please write your own downloading logic.
    WARNING:tensorflow:From /usr/local/lib/python3.6/dist-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:262: extract_images (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use tf.data to implement this functionality.
    Extracting MNIST_data/train-images-idx3-ubyte.gz
    WARNING:tensorflow:From /usr/local/lib/python3.6/dist-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:267: extract_labels (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use tf.data to implement this functionality.
    Extracting MNIST_data/train-labels-idx1-ubyte.gz
    WARNING:tensorflow:From /usr/local/lib/python3.6/dist-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:110: dense_to_one_hot (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use tf.one_hot on tensors.
    Extracting MNIST_data/t10k-images-idx3-ubyte.gz
    Extracting MNIST_data/t10k-labels-idx1-ubyte.gz
    WARNING:tensorflow:From /usr/local/lib/python3.6/dist-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:290: DataSet.__init__ (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use alternatives such as official/mnist/dataset.py from tensorflow/models.
    WARNING:tensorflow:From /usr/local/lib/python3.6/dist-packages/tensorflow/python/util/tf_should_use.py:189: initialize_all_variables (from tensorflow.python.ops.variables) is deprecated and will be removed after 2017-03-02.
    Instructions for updating:
    Use `tf.global_variables_initializer` instead.
    0.9163


Wrap model using s2i


```python
!s2i build . seldonio/seldon-core-s2i-python36:0.7 deep-mnist:0.1
```

    ---> Installing application source...
    ---> Installing dependencies ...
    Looking in links: /whl
    Requirement already satisfied: tensorflow>=1.12.0 in /usr/local/lib/python3.6/site-packages (from -r requirements.txt (line 1)) (1.13.1)
    Requirement already satisfied: grpcio>=1.8.6 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.20.1)
    Requirement already satisfied: six>=1.10.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.12.0)
    Requirement already satisfied: absl-py>=0.1.6 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.7.1)
    Requirement already satisfied: wheel>=0.26 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.32.2)
    Requirement already satisfied: tensorflow-estimator<1.14.0rc0,>=1.13.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.13.0)
    Requirement already satisfied: astor>=0.6.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.7.1)
    Requirement already satisfied: keras-preprocessing>=1.0.5 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.9)
    Requirement already satisfied: keras-applications>=1.0.6 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.7)
    Requirement already satisfied: protobuf>=3.6.1 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.7.1)
    Requirement already satisfied: numpy>=1.13.3 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.16.3)
    Requirement already satisfied: tensorboard<1.14.0,>=1.13.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.13.1)
    Requirement already satisfied: termcolor>=1.1.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.1.0)
    Requirement already satisfied: gast>=0.2.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.2.2)
    Requirement already satisfied: mock>=2.0.0 in /usr/local/lib/python3.6/site-packages (from tensorflow-estimator<1.14.0rc0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.0.5)
    Requirement already satisfied: h5py in /usr/local/lib/python3.6/site-packages (from keras-applications>=1.0.6->tensorflow>=1.12.0->-r requirements.txt (line 1)) (2.9.0)
    Requirement already satisfied: setuptools in /usr/local/lib/python3.6/site-packages (from protobuf>=3.6.1->tensorflow>=1.12.0->-r requirements.txt (line 1)) (40.6.2)
    Requirement already satisfied: markdown>=2.6.8 in /usr/local/lib/python3.6/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.1)
    Requirement already satisfied: werkzeug>=0.11.15 in /usr/local/lib/python3.6/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.15.4)
    Url '/whl' is ignored. It is either a non-existing path or lacks a specific scheme.
    You are using pip version 18.1, however version 19.1.1 is available.
    You should consider upgrading via the 'pip install --upgrade pip' command.
    Build completed successfully



```python
!docker run --name "mnist_predictor" -d --rm -p 5000:5000 deep-mnist:0.1
```

    97da1e904b4a24cf2e6aaf4a3f9dc15cbc376b0e02af6521a5d042621d19d7fe


Send some random features that conform to the contract


```python
!seldon-core-tester contract.json 0.0.0.0 5000 -p
```

    ----------------------------------------
    SENDING NEW REQUEST:
    {'meta': {}, 'data': {'names': [u'x1', u'x2', u'x3', u'x4', u'x5', u'x6', u'x7', u'x8', u'x9', u'x10', u'x11', u'x12', u'x13', u'x14', u'x15', u'x16', u'x17', u'x18', u'x19', u'x20', u'x21', u'x22', u'x23', u'x24', u'x25', u'x26', u'x27', u'x28', u'x29', u'x30', u'x31', u'x32', u'x33', u'x34', u'x35', u'x36', u'x37', u'x38', u'x39', u'x40', u'x41', u'x42', u'x43', u'x44', u'x45', u'x46', u'x47', u'x48', u'x49', u'x50', u'x51', u'x52', u'x53', u'x54', u'x55', u'x56', u'x57', u'x58', u'x59', u'x60', u'x61', u'x62', u'x63', u'x64', u'x65', u'x66', u'x67', u'x68', u'x69', u'x70', u'x71', u'x72', u'x73', u'x74', u'x75', u'x76', u'x77', u'x78', u'x79', u'x80', u'x81', u'x82', u'x83', u'x84', u'x85', u'x86', u'x87', u'x88', u'x89', u'x90', u'x91', u'x92', u'x93', u'x94', u'x95', u'x96', u'x97', u'x98', u'x99', u'x100', u'x101', u'x102', u'x103', u'x104', u'x105', u'x106', u'x107', u'x108', u'x109', u'x110', u'x111', u'x112', u'x113', u'x114', u'x115', u'x116', u'x117', u'x118', u'x119', u'x120', u'x121', u'x122', u'x123', u'x124', u'x125', u'x126', u'x127', u'x128', u'x129', u'x130', u'x131', u'x132', u'x133', u'x134', u'x135', u'x136', u'x137', u'x138', u'x139', u'x140', u'x141', u'x142', u'x143', u'x144', u'x145', u'x146', u'x147', u'x148', u'x149', u'x150', u'x151', u'x152', u'x153', u'x154', u'x155', u'x156', u'x157', u'x158', u'x159', u'x160', u'x161', u'x162', u'x163', u'x164', u'x165', u'x166', u'x167', u'x168', u'x169', u'x170', u'x171', u'x172', u'x173', u'x174', u'x175', u'x176', u'x177', u'x178', u'x179', u'x180', u'x181', u'x182', u'x183', u'x184', u'x185', u'x186', u'x187', u'x188', u'x189', u'x190', u'x191', u'x192', u'x193', u'x194', u'x195', u'x196', u'x197', u'x198', u'x199', u'x200', u'x201', u'x202', u'x203', u'x204', u'x205', u'x206', u'x207', u'x208', u'x209', u'x210', u'x211', u'x212', u'x213', u'x214', u'x215', u'x216', u'x217', u'x218', u'x219', u'x220', u'x221', u'x222', u'x223', u'x224', u'x225', u'x226', u'x227', u'x228', u'x229', u'x230', u'x231', u'x232', u'x233', u'x234', u'x235', u'x236', u'x237', u'x238', u'x239', u'x240', u'x241', u'x242', u'x243', u'x244', u'x245', u'x246', u'x247', u'x248', u'x249', u'x250', u'x251', u'x252', u'x253', u'x254', u'x255', u'x256', u'x257', u'x258', u'x259', u'x260', u'x261', u'x262', u'x263', u'x264', u'x265', u'x266', u'x267', u'x268', u'x269', u'x270', u'x271', u'x272', u'x273', u'x274', u'x275', u'x276', u'x277', u'x278', u'x279', u'x280', u'x281', u'x282', u'x283', u'x284', u'x285', u'x286', u'x287', u'x288', u'x289', u'x290', u'x291', u'x292', u'x293', u'x294', u'x295', u'x296', u'x297', u'x298', u'x299', u'x300', u'x301', u'x302', u'x303', u'x304', u'x305', u'x306', u'x307', u'x308', u'x309', u'x310', u'x311', u'x312', u'x313', u'x314', u'x315', u'x316', u'x317', u'x318', u'x319', u'x320', u'x321', u'x322', u'x323', u'x324', u'x325', u'x326', u'x327', u'x328', u'x329', u'x330', u'x331', u'x332', u'x333', u'x334', u'x335', u'x336', u'x337', u'x338', u'x339', u'x340', u'x341', u'x342', u'x343', u'x344', u'x345', u'x346', u'x347', u'x348', u'x349', u'x350', u'x351', u'x352', u'x353', u'x354', u'x355', u'x356', u'x357', u'x358', u'x359', u'x360', u'x361', u'x362', u'x363', u'x364', u'x365', u'x366', u'x367', u'x368', u'x369', u'x370', u'x371', u'x372', u'x373', u'x374', u'x375', u'x376', u'x377', u'x378', u'x379', u'x380', u'x381', u'x382', u'x383', u'x384', u'x385', u'x386', u'x387', u'x388', u'x389', u'x390', u'x391', u'x392', u'x393', u'x394', u'x395', u'x396', u'x397', u'x398', u'x399', u'x400', u'x401', u'x402', u'x403', u'x404', u'x405', u'x406', u'x407', u'x408', u'x409', u'x410', u'x411', u'x412', u'x413', u'x414', u'x415', u'x416', u'x417', u'x418', u'x419', u'x420', u'x421', u'x422', u'x423', u'x424', u'x425', u'x426', u'x427', u'x428', u'x429', u'x430', u'x431', u'x432', u'x433', u'x434', u'x435', u'x436', u'x437', u'x438', u'x439', u'x440', u'x441', u'x442', u'x443', u'x444', u'x445', u'x446', u'x447', u'x448', u'x449', u'x450', u'x451', u'x452', u'x453', u'x454', u'x455', u'x456', u'x457', u'x458', u'x459', u'x460', u'x461', u'x462', u'x463', u'x464', u'x465', u'x466', u'x467', u'x468', u'x469', u'x470', u'x471', u'x472', u'x473', u'x474', u'x475', u'x476', u'x477', u'x478', u'x479', u'x480', u'x481', u'x482', u'x483', u'x484', u'x485', u'x486', u'x487', u'x488', u'x489', u'x490', u'x491', u'x492', u'x493', u'x494', u'x495', u'x496', u'x497', u'x498', u'x499', u'x500', u'x501', u'x502', u'x503', u'x504', u'x505', u'x506', u'x507', u'x508', u'x509', u'x510', u'x511', u'x512', u'x513', u'x514', u'x515', u'x516', u'x517', u'x518', u'x519', u'x520', u'x521', u'x522', u'x523', u'x524', u'x525', u'x526', u'x527', u'x528', u'x529', u'x530', u'x531', u'x532', u'x533', u'x534', u'x535', u'x536', u'x537', u'x538', u'x539', u'x540', u'x541', u'x542', u'x543', u'x544', u'x545', u'x546', u'x547', u'x548', u'x549', u'x550', u'x551', u'x552', u'x553', u'x554', u'x555', u'x556', u'x557', u'x558', u'x559', u'x560', u'x561', u'x562', u'x563', u'x564', u'x565', u'x566', u'x567', u'x568', u'x569', u'x570', u'x571', u'x572', u'x573', u'x574', u'x575', u'x576', u'x577', u'x578', u'x579', u'x580', u'x581', u'x582', u'x583', u'x584', u'x585', u'x586', u'x587', u'x588', u'x589', u'x590', u'x591', u'x592', u'x593', u'x594', u'x595', u'x596', u'x597', u'x598', u'x599', u'x600', u'x601', u'x602', u'x603', u'x604', u'x605', u'x606', u'x607', u'x608', u'x609', u'x610', u'x611', u'x612', u'x613', u'x614', u'x615', u'x616', u'x617', u'x618', u'x619', u'x620', u'x621', u'x622', u'x623', u'x624', u'x625', u'x626', u'x627', u'x628', u'x629', u'x630', u'x631', u'x632', u'x633', u'x634', u'x635', u'x636', u'x637', u'x638', u'x639', u'x640', u'x641', u'x642', u'x643', u'x644', u'x645', u'x646', u'x647', u'x648', u'x649', u'x650', u'x651', u'x652', u'x653', u'x654', u'x655', u'x656', u'x657', u'x658', u'x659', u'x660', u'x661', u'x662', u'x663', u'x664', u'x665', u'x666', u'x667', u'x668', u'x669', u'x670', u'x671', u'x672', u'x673', u'x674', u'x675', u'x676', u'x677', u'x678', u'x679', u'x680', u'x681', u'x682', u'x683', u'x684', u'x685', u'x686', u'x687', u'x688', u'x689', u'x690', u'x691', u'x692', u'x693', u'x694', u'x695', u'x696', u'x697', u'x698', u'x699', u'x700', u'x701', u'x702', u'x703', u'x704', u'x705', u'x706', u'x707', u'x708', u'x709', u'x710', u'x711', u'x712', u'x713', u'x714', u'x715', u'x716', u'x717', u'x718', u'x719', u'x720', u'x721', u'x722', u'x723', u'x724', u'x725', u'x726', u'x727', u'x728', u'x729', u'x730', u'x731', u'x732', u'x733', u'x734', u'x735', u'x736', u'x737', u'x738', u'x739', u'x740', u'x741', u'x742', u'x743', u'x744', u'x745', u'x746', u'x747', u'x748', u'x749', u'x750', u'x751', u'x752', u'x753', u'x754', u'x755', u'x756', u'x757', u'x758', u'x759', u'x760', u'x761', u'x762', u'x763', u'x764', u'x765', u'x766', u'x767', u'x768', u'x769', u'x770', u'x771', u'x772', u'x773', u'x774', u'x775', u'x776', u'x777', u'x778', u'x779', u'x780', u'x781', u'x782', u'x783', u'x784'], 'ndarray': [[0.187, 0.807, 0.517, 0.695, 0.122, 0.597, 0.139, 0.102, 0.379, 0.355, 0.483, 0.53, 0.469, 0.917, 0.449, 0.89, 0.775, 0.296, 0.371, 0.299, 0.462, 0.937, 0.456, 0.741, 0.375, 0.14, 0.819, 0.079, 0.344, 0.252, 0.384, 0.79, 0.233, 0.496, 0.444, 0.911, 0.803, 0.806, 0.226, 0.242, 0.46, 0.307, 0.045, 0.638, 0.798, 0.82, 0.427, 0.154, 0.84, 0.86, 0.437, 0.613, 0.509, 0.161, 0.883, 0.333, 0.424, 0.092, 0.19, 0.025, 0.907, 0.126, 0.355, 0.436, 0.117, 0.585, 0.383, 0.886, 0.48, 0.701, 0.498, 0.294, 0.145, 0.723, 0.746, 0.883, 0.549, 0.557, 0.108, 0.852, 0.918, 0.554, 0.531, 0.721, 0.456, 0.461, 0.674, 0.225, 0.414, 0.197, 0.775, 0.598, 0.126, 0.936, 0.823, 0.601, 0.706, 0.654, 0.662, 0.282, 0.634, 0.72, 0.062, 0.035, 0.092, 0.443, 0.44, 0.793, 0.393, 0.436, 0.184, 0.68, 0.251, 0.171, 0.647, 0.512, 0.17, 0.247, 0.693, 0.745, 0.14, 0.713, 0.189, 0.543, 0.301, 0.673, 0.252, 0.054, 0.754, 0.533, 0.572, 0.526, 0.982, 0.017, 0.812, 0.006, 0.771, 0.516, 0.246, 0.505, 0.736, 0.41, 0.966, 0.905, 0.424, 0.941, 0.318, 0.943, 0.867, 0.014, 0.921, 0.123, 0.644, 0.498, 0.871, 0.449, 0.887, 0.059, 0.536, 0.675, 0.488, 0.514, 0.964, 0.537, 0.682, 0.83, 0.386, 0.019, 0.582, 0.13, 0.043, 0.804, 0.087, 0.031, 0.661, 0.637, 0.333, 0.426, 0.184, 0.77, 0.607, 0.915, 0.878, 0.005, 0.399, 0.754, 0.432, 0.652, 0.944, 0.252, 0.522, 0.653, 0.465, 0.461, 0.785, 0.531, 0.265, 0.115, 0.52, 0.612, 0.899, 0.668, 0.169, 0.931, 0.968, 0.116, 0.587, 0.949, 0.009, 0.669, 0.683, 0.29, 0.462, 0.306, 0.321, 0.64, 0.517, 0.849, 0.326, 0.106, 0.892, 0.76, 0.853, 0.574, 0.298, 0.138, 0.866, 0.848, 0.382, 0.588, 0.634, 0.814, 0.581, 0.19, 0.991, 0.012, 0.627, 0.668, 0.066, 0.17, 0.017, 0.885, 0.885, 0.591, 0.782, 0.265, 0.574, 0.774, 0.551, 0.852, 0.217, 0.884, 0.447, 0.76, 0.347, 0.589, 0.649, 0.983, 0.135, 0.383, 0.077, 0.914, 0.228, 0.147, 0.34, 0.521, 0.295, 0.291, 0.001, 0.79, 0.372, 0.822, 0.153, 0.315, 0.395, 0.01, 0.94, 0.29, 0.406, 0.584, 0.304, 0.884, 0.759, 0.17, 0.4, 0.027, 0.111, 0.219, 0.48, 0.832, 0.164, 0.344, 0.113, 0.443, 0.847, 0.217, 0.806, 0.713, 0.534, 0.898, 0.93, 0.959, 0.75, 0.115, 0.171, 0.535, 0.25, 0.281, 0.724, 0.982, 0.053, 0.116, 0.002, 0.438, 0.557, 0.604, 0.628, 0.64, 0.99, 0.705, 0.125, 0.918, 0.653, 0.934, 0.796, 0.678, 0.352, 0.574, 0.333, 0.297, 0.425, 0.8, 0.785, 0.081, 0.845, 0.425, 0.531, 0.41, 0.606, 0.631, 0.255, 0.118, 0.647, 0.003, 0.799, 0.309, 0.439, 0.86, 0.751, 0.759, 0.092, 0.819, 0.188, 0.89, 0.873, 0.624, 0.501, 0.494, 0.252, 0.976, 0.621, 0.378, 0.473, 0.272, 0.496, 0.36, 0.814, 0.521, 0.23, 0.262, 0.132, 0.308, 0.079, 0.347, 0.358, 0.99, 0.506, 0.619, 0.151, 0.999, 0.485, 0.229, 0.933, 0.118, 0.823, 0.079, 0.496, 0.26, 0.432, 0.218, 0.93, 0.954, 0.998, 0.549, 0.547, 0.596, 0.369, 0.008, 0.226, 0.611, 0.429, 0.133, 0.026, 0.234, 0.619, 0.523, 0.062, 0.434, 0.343, 0.202, 0.857, 0.219, 0.363, 0.282, 0.054, 0.905, 0.883, 0.376, 0.182, 0.274, 0.752, 0.345, 0.627, 0.272, 0.509, 0.149, 0.642, 0.9, 0.24, 0.341, 0.485, 0.819, 0.976, 0.951, 0.081, 0.444, 0.286, 0.19, 0.659, 0.299, 0.228, 0.989, 0.692, 0.882, 0.29, 0.88, 0.295, 0.716, 0.74, 0.893, 0.827, 0.125, 0.952, 0.804, 0.168, 0.494, 0.696, 0.991, 0.519, 0.275, 0.043, 0.738, 0.533, 0.517, 0.249, 0.699, 0.258, 0.407, 0.435, 0.523, 0.14, 0.245, 0.82, 0.59, 0.426, 0.005, 0.036, 0.183, 0.166, 0.621, 0.41, 0.459, 0.374, 0.931, 0.665, 0.369, 0.089, 0.461, 0.727, 0.229, 0.783, 0.5, 0.219, 0.634, 0.74, 0.233, 0.101, 0.881, 0.609, 0.928, 0.25, 0.814, 0.866, 0.198, 0.26, 0.945, 0.989, 0.714, 0.696, 0.416, 0.943, 0.262, 0.174, 0.158, 0.335, 0.808, 0.712, 0.528, 0.848, 0.138, 0.572, 0.089, 0.549, 0.724, 0.825, 0.624, 0.361, 0.979, 0.498, 0.761, 0.498, 0.635, 0.456, 0.315, 0.623, 0.924, 0.252, 0.134, 0.503, 0.472, 0.275, 0.032, 0.375, 0.767, 0.888, 0.811, 0.761, 0.781, 0.983, 0.43, 0.156, 0.565, 0.427, 0.199, 0.547, 0.573, 0.288, 0.758, 0.289, 0.38, 0.78, 0.988, 0.614, 0.959, 0.808, 0.752, 0.589, 0.501, 0.64, 0.646, 0.598, 0.487, 0.577, 0.618, 0.153, 0.34, 0.86, 0.034, 0.673, 0.266, 0.053, 0.684, 0.104, 0.36, 0.302, 0.774, 0.732, 0.9, 0.968, 0.698, 0.419, 0.249, 0.164, 0.466, 0.698, 0.882, 0.215, 0.634, 0.284, 0.953, 0.303, 0.451, 0.138, 0.522, 0.345, 0.01, 0.185, 0.101, 0.556, 0.851, 0.135, 0.756, 0.424, 0.373, 0.266, 0.088, 0.121, 0.115, 0.702, 0.259, 0.398, 0.771, 0.975, 0.905, 0.832, 0.369, 0.934, 0.54, 0.186, 0.455, 0.167, 0.908, 0.793, 0.766, 0.306, 0.707, 0.354, 0.599, 0.191, 0.727, 0.059, 0.013, 0.623, 0.072, 0.648, 0.684, 0.182, 0.143, 0.287, 0.018, 0.274, 0.933, 0.092, 0.981, 0.832, 0.105, 0.545, 0.287, 0.247, 0.031, 0.988, 0.065, 0.707, 0.523, 0.46, 0.333, 0.756, 0.697, 0.225, 0.5, 0.929, 0.44, 0.356, 0.924, 0.761, 0.788, 0.644, 0.872, 0.084, 0.242, 0.068, 0.578, 0.588, 0.584, 0.308, 0.23, 0.222, 0.271, 0.099, 0.556, 0.7, 0.579, 0.993, 0.052, 0.856, 0.291, 0.528, 0.593, 0.802, 0.701, 0.597, 0.762, 0.869, 0.831, 0.272, 0.399, 0.138, 0.569, 0.021, 0.795, 0.807, 0.281, 0.852, 0.442, 0.557, 0.453, 0.613, 0.319, 0.55, 0.541, 0.886, 0.351, 0.642, 0.002, 0.789, 0.78, 0.019, 0.866, 0.187, 0.582, 0.439, 0.491, 0.043, 0.911, 0.903, 0.746, 0.462, 0.716, 0.185, 0.854, 0.647, 0.193, 0.587, 0.212, 0.395, 0.226, 0.24, 0.643, 0.325, 0.16, 0.376, 0.799, 0.445, 0.429, 0.462, 0.127, 0.698, 0.785, 0.624, 0.091, 0.193, 0.73, 0.012, 0.73, 0.887, 0.53, 0.906, 0.627, 0.134, 0.331, 0.301, 0.59, 0.413, 0.101, 0.453, 0.767, 0.807, 0.358, 0.673, 0.679, 0.164, 0.39, 0.423, 0.89]]}}
    RECEIVED RESPONSE:
    {u'meta': {}, u'data': {u'names': [u'class:0', u'class:1', u'class:2', u'class:3', u'class:4', u'class:5', u'class:6', u'class:7', u'class:8', u'class:9'], u'ndarray': [[0.0009108927333727479, 5.9436615629238077e-08, 0.27207350730895996, 0.490738183259964, 2.75773504654353e-06, 0.20564192533493042, 9.108992526307702e-05, 0.0001797660515876487, 0.030353713780641556, 8.07244850875577e-06]]}}
    ()
    Time 0.18249297142



```python
!docker rm mnist_predictor --force
```

    mnist_predictor


## Test using Minikube

**Due to a [minikube/s2i issue](https://github.com/SeldonIO/seldon-core/issues/253) you will need [s2i >= 1.1.13](https://github.com/openshift/source-to-image/releases/tag/v1.1.13)**


```python
!minikube start --memory 4096
```

    😄  minikube v0.34.1 on linux (amd64)
    🔥  Creating virtualbox VM (CPUs=2, Memory=4096MB, Disk=20000MB) ...
    📶  "minikube" IP address is 192.168.99.100
    🐳  Configuring Docker as the container runtime ...
    ✨  Preparing Kubernetes environment ...
    🚜  Pulling images required by Kubernetes v1.13.3 ...
    🚀  Launching Kubernetes v1.13.3 using kubeadm ... 
    🔑  Configuring cluster permissions ...
    🤔  Verifying component health .....
    💗  kubectl is now configured to use "minikube"
    🏄  Done! Thank you for using minikube!



```python
!kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

    clusterrolebinding.rbac.authorization.k8s.io/kube-system-cluster-admin created



```python
!helm init
```

    $HELM_HOME has been configured at /home/clive/.helm.
    
    Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
    
    Please note: by default, Tiller is deployed with an insecure 'allow unauthenticated users' policy.
    To prevent this, run `helm init` with the --tiller-tls-verify flag.
    For more information on securing your installation see: https://docs.helm.sh/using_helm/#securing-your-helm-installation
    Happy Helming!



```python
!kubectl rollout status deploy/tiller-deploy -n kube-system
```

    Waiting for deployment "tiller-deploy" rollout to finish: 0 of 1 updated replicas are available...
    deployment "tiller-deploy" successfully rolled out



```python
!helm install ../../../helm-charts/seldon-core-operator --name seldon-core --set usageMetrics.enabled=true --namespace seldon-system
```

    NAME:   seldon-core
    LAST DEPLOYED: Thu Apr 25 09:13:58 2019
    NAMESPACE: seldon-system
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1beta1/CustomResourceDefinition
    NAME                                         AGE
    seldondeployments.machinelearning.seldon.io  0s
    
    ==> v1/ClusterRole
    seldon-operator-manager-role  0s
    
    ==> v1/ClusterRoleBinding
    NAME                                 AGE
    seldon-operator-manager-rolebinding  0s
    
    ==> v1/Service
    NAME                                        TYPE       CLUSTER-IP    EXTERNAL-IP  PORT(S)  AGE
    seldon-operator-controller-manager-service  ClusterIP  10.109.84.44  <none>       443/TCP  0s
    
    ==> v1/StatefulSet
    NAME                                DESIRED  CURRENT  AGE
    seldon-operator-controller-manager  1        1        0s
    
    ==> v1/Pod(related)
    NAME                                  READY  STATUS             RESTARTS  AGE
    seldon-operator-controller-manager-0  0/1    ContainerCreating  0         0s
    
    ==> v1/Secret
    NAME                                   TYPE    DATA  AGE
    seldon-operator-webhook-server-secret  Opaque  0     0s
    
    
    NOTES:
    NOTES: TODO
    
    



```python
!kubectl rollout status statefulset.apps/seldon-operator-controller-manager -n seldon-system
```

    partitioned roll out complete: 1 new pods have been updated...


## Setup Ingress
There are gRPC issues with the latest Ambassador, so we rewcommend 0.40.2 until these are fixed.


```python
!helm install stable/ambassador --name ambassador --set crds.keep=false
```

    NAME:   ambassador
    LAST DEPLOYED: Thu Apr 25 09:14:31 2019
    NAMESPACE: default
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1/ServiceAccount
    NAME        SECRETS  AGE
    ambassador  1        0s
    
    ==> v1beta1/ClusterRole
    NAME        AGE
    ambassador  0s
    
    ==> v1beta1/ClusterRoleBinding
    NAME        AGE
    ambassador  0s
    
    ==> v1/Service
    NAME               TYPE          CLUSTER-IP     EXTERNAL-IP  PORT(S)                     AGE
    ambassador-admins  ClusterIP     10.110.99.128  <none>       8877/TCP                    0s
    ambassador         LoadBalancer  10.97.7.72     <pending>    80:30064/TCP,443:32402/TCP  0s
    
    ==> v1/Deployment
    NAME        DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
    ambassador  3        3        3           0          0s
    
    ==> v1/Pod(related)
    NAME                         READY  STATUS             RESTARTS  AGE
    ambassador-5b89d44544-5hhl7  0/1    ContainerCreating  0         0s
    ambassador-5b89d44544-5xcdw  0/1    ContainerCreating  0         0s
    ambassador-5b89d44544-7rv6r  0/1    ContainerCreating  0         0s
    
    
    NOTES:
    Congratuations! You've successfully installed Ambassador.
    
    For help, visit our Slack at https://d6e.co/slack or view the documentation online at https://www.getambassador.io.
    
    To get the IP address of Ambassador, run the following commands:
    NOTE: It may take a few minutes for the LoadBalancer IP to be available.
         You can watch the status of by running 'kubectl get svc -w  --namespace default ambassador'
    
      On GKE/Azure:
      export SERVICE_IP=$(kubectl get svc --namespace default ambassador -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
      On AWS:
      export SERVICE_IP=$(kubectl get svc --namespace default ambassador -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    
      echo http://$SERVICE_IP:
    



```python
!kubectl rollout status deployment.apps/ambassador
```

    Waiting for deployment "ambassador" rollout to finish: 0 of 3 updated replicas are available...
    Waiting for deployment "ambassador" rollout to finish: 1 of 3 updated replicas are available...
    Waiting for deployment "ambassador" rollout to finish: 2 of 3 updated replicas are available...
    deployment "ambassador" successfully rolled out


## Wrap Model and Test


```python
!eval $(minikube docker-env) && s2i build . seldonio/seldon-core-s2i-python2:0.5.1 deep-mnist:0.1
```

    ---> Installing application source...
    ---> Installing dependencies ...
    DEPRECATION: Python 2.7 will reach the end of its life on January 1st, 2020. Please upgrade your Python as Python 2.7 won't be maintained after that date. A future version of pip will drop support for Python 2.7.
    Looking in links: /whl
    Requirement already satisfied: tensorflow>=1.12.0 in /usr/local/lib/python2.7/site-packages (from -r requirements.txt (line 1)) (1.13.1)
    Requirement already satisfied: astor>=0.6.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.7.1)
    Requirement already satisfied: keras-preprocessing>=1.0.5 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.9)
    Requirement already satisfied: gast>=0.2.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.2.2)
    Requirement already satisfied: enum34>=1.1.6 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.1.6)
    Requirement already satisfied: protobuf>=3.6.1 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.7.0)
    Requirement already satisfied: six>=1.10.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.12.0)
    Requirement already satisfied: absl-py>=0.1.6 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.7.1)
    Requirement already satisfied: backports.weakref>=1.0rc1 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.post1)
    Requirement already satisfied: tensorboard<1.14.0,>=1.13.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.13.1)
    Requirement already satisfied: wheel in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.33.1)
    Requirement already satisfied: termcolor>=1.1.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.1.0)
    Requirement already satisfied: numpy>=1.13.3 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.16.2)
    Requirement already satisfied: mock>=2.0.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (2.0.0)
    Requirement already satisfied: keras-applications>=1.0.6 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.7)
    Requirement already satisfied: tensorflow-estimator<1.14.0rc0,>=1.13.0 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.13.0)
    Requirement already satisfied: grpcio>=1.8.6 in /usr/local/lib/python2.7/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.19.0)
    Requirement already satisfied: setuptools in /usr/local/lib/python2.7/site-packages (from protobuf>=3.6.1->tensorflow>=1.12.0->-r requirements.txt (line 1)) (40.8.0)
    Requirement already satisfied: markdown>=2.6.8 in /usr/local/lib/python2.7/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.0.1)
    Requirement already satisfied: futures>=3.1.1; python_version < "3" in /usr/local/lib/python2.7/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.2.0)
    Requirement already satisfied: werkzeug>=0.11.15 in /usr/local/lib/python2.7/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.15.0)
    Requirement already satisfied: funcsigs>=1; python_version < "3.3" in /usr/local/lib/python2.7/site-packages (from mock>=2.0.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.2)
    Requirement already satisfied: pbr>=0.11 in /usr/local/lib/python2.7/site-packages (from mock>=2.0.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (5.1.3)
    Requirement already satisfied: h5py in /usr/local/lib/python2.7/site-packages (from keras-applications>=1.0.6->tensorflow>=1.12.0->-r requirements.txt (line 1)) (2.9.0)
    Url '/whl' is ignored. It is either a non-existing path or lacks a specific scheme.
    You are using pip version 19.0.3, however version 19.1 is available.
    You should consider upgrading via the 'pip install --upgrade pip' command.
    Build completed successfully



```python
!kubectl create -f deep_mnist.json
```

    seldondeployment.machinelearning.seldon.io/deep-mnist created



```python
!kubectl rollout status deploy/deep-mnist-single-model-8969cc0
```

    Waiting for deployment "deep-mnist-single-model-8969cc0" rollout to finish: 0 of 1 updated replicas are available...
    deployment "deep-mnist-single-model-8969cc0" successfully rolled out



```python
!seldon-core-api-tester contract.json `minikube ip` `kubectl get svc ambassador -o jsonpath='{.spec.ports[0].nodePort}'` \
    deep-mnist --namespace default -p
```

    ----------------------------------------
    SENDING NEW REQUEST:
    
    [[0.07  0.525 0.195 0.946 0.425 0.312 0.099 0.855 0.955 0.769 0.156 0.647
      0.479 0.197 0.586 0.616 0.105 0.862 0.073 0.335 0.277 0.345 0.872 0.247
      0.266 0.289 0.396 0.217 0.143 0.685 0.567 0.425 0.919 0.474 0.436 0.6
      0.341 0.776 0.417 0.541 0.62  0.161 0.164 0.757 0.135 0.982 0.491 0.735
      0.837 0.387 0.628 0.069 0.062 0.73  0.742 0.563 0.22  0.964 0.01  0.084
      0.681 0.553 0.746 0.834 0.143 0.34  0.676 0.794 0.562 0.113 0.195 0.309
      0.334 0.45  0.936 0.233 0.435 0.105 0.347 0.149 0.378 0.939 0.844 0.912
      0.869 0.251 0.231 0.596 0.603 0.716 0.086 0.669 0.78  0.265 0.316 0.063
      0.296 0.347 0.23  0.843 0.031 0.923 0.978 0.623 0.738 0.362 0.186 0.905
      0.138 0.952 0.209 0.218 0.407 0.198 0.489 0.838 0.372 0.335 0.908 0.505
      0.551 0.256 0.966 0.827 0.121 0.642 0.321 0.949 0.225 0.903 0.954 0.193
      0.378 0.109 0.684 0.026 0.804 0.108 0.104 0.646 0.101 0.097 0.303 0.528
      0.49  0.91  0.523 0.868 0.22  0.555 0.353 0.627 0.077 0.946 0.127 0.101
      0.341 0.205 0.004 0.963 0.825 0.699 0.222 0.644 0.895 0.219 0.151 0.682
      0.488 0.78  0.443 0.8   0.527 0.524 0.894 0.797 0.192 0.744 0.096 0.222
      0.953 0.219 0.244 0.335 0.932 0.507 0.613 0.911 0.501 0.548 0.168 0.27
      0.998 0.889 0.866 0.406 0.042 0.159 0.938 0.94  0.549 0.229 0.965 0.392
      0.943 0.656 0.822 0.336 0.432 0.176 0.726 0.142 0.696 0.899 0.325 0.596
      0.422 0.036 0.381 0.407 0.943 0.249 0.963 0.652 0.226 0.333 0.207 0.825
      0.611 0.752 0.196 0.452 0.616 0.146 0.02  0.804 0.466 0.792 0.241 0.861
      0.762 0.606 0.721 0.404 0.95  0.044 0.911 0.424 0.19  0.14  0.756 0.982
      0.487 0.008 0.209 0.922 0.211 0.29  0.966 0.996 0.097 0.308 0.944 0.054
      0.439 0.522 0.362 0.497 0.943 0.338 0.233 0.471 0.7   0.396 0.598 0.713
      0.708 0.886 0.118 0.615 0.946 0.066 0.069 0.046 0.414 0.298 0.988 0.7
      0.396 0.685 0.521 0.495 0.523 0.596 0.606 0.364 0.937 0.023 0.396 0.565
      0.276 0.034 0.243 0.42  0.222 0.687 0.364 0.111 0.205 0.69  0.344 0.497
      0.881 0.094 0.921 0.137 0.379 0.347 0.161 0.53  0.758 0.215 0.322 0.559
      0.249 0.751 0.991 0.966 0.333 0.44  0.912 0.863 0.666 0.495 0.808 0.932
      0.191 0.279 0.317 0.241 0.678 0.735 0.092 0.751 0.356 0.435 0.33  0.153
      0.232 0.265 0.307 0.12  0.121 0.422 0.283 0.039 0.024 0.097 0.33  0.67
      0.917 0.519 0.423 0.24  0.168 0.466 0.288 0.777 0.509 0.055 0.211 0.382
      0.329 0.394 0.391 0.122 0.284 0.751 0.345 0.003 0.308 0.222 0.234 0.389
      0.062 0.733 0.358 0.804 0.377 0.598 0.293 0.096 0.316 0.798 0.1   0.632
      0.55  0.36  0.157 0.211 0.813 0.897 0.598 0.78  0.134 0.548 0.284 0.84
      0.447 0.131 0.178 0.316 0.527 0.271 0.437 0.72  0.096 0.613 0.532 0.323
      0.17  0.701 0.84  0.155 0.737 0.471 0.407 0.979 0.58  0.694 0.611 0.276
      0.113 0.084 0.024 0.18  0.709 0.716 0.469 0.804 0.483 0.307 0.055 0.226
      0.377 0.297 0.56  0.021 0.581 0.541 0.471 0.205 0.6   0.828 0.794 0.748
      0.277 0.635 0.3   0.571 0.577 0.193 0.204 0.244 0.408 0.341 0.626 0.434
      0.502 0.585 0.107 0.816 0.928 0.612 0.286 0.983 0.178 0.703 0.978 0.208
      0.5   0.424 0.384 0.015 0.418 0.339 0.043 0.699 0.533 0.625 0.834 0.266
      0.336 0.029 0.718 0.074 0.252 0.018 0.331 0.882 0.591 0.364 0.008 0.415
      0.271 0.962 0.144 0.939 0.858 0.258 0.688 0.401 0.03  0.432 0.823 0.69
      0.824 0.284 0.971 0.022 0.47  0.482 0.938 0.201 0.635 0.612 0.975 0.929
      0.478 0.023 0.968 0.63  0.605 0.26  0.416 0.039 0.583 0.538 0.167 0.374
      0.694 0.128 0.692 0.786 0.664 0.343 0.53  0.207 0.217 0.691 0.239 0.121
      0.072 0.806 0.72  0.069 0.799 0.789 0.058 0.889 0.657 0.168 0.18  0.337
      0.48  0.471 0.16  0.44  0.733 0.699 0.439 0.006 0.681 0.177 0.366 0.515
      0.415 0.927 0.26  0.121 0.794 0.257 0.837 0.51  0.45  0.41  0.09  0.017
      0.856 0.06  0.341 1.    0.424 0.892 0.276 0.216 0.52  0.755 0.965 0.757
      0.37  0.204 0.456 0.306 0.72  0.233 0.289 0.359 0.478 0.063 0.249 0.816
      0.568 0.978 0.191 0.588 0.872 0.783 0.76  0.696 0.305 0.832 0.173 0.515
      0.459 0.471 0.386 0.825 0.625 0.495 0.596 0.426 0.159 0.174 0.519 0.355
      0.799 0.98  0.606 0.797 0.81  0.111 0.888 0.583 0.163 0.907 0.336 0.708
      0.815 0.171 0.454 0.359 0.19  0.775 0.488 0.674 0.905 0.889 0.606 0.429
      0.387 0.724 0.204 0.145 0.649 0.306 0.811 0.325 0.022 0.573 0.881 0.474
      0.413 0.981 0.074 0.898 0.715 0.323 0.942 0.586 0.857 0.03  0.997 0.72
      0.908 0.332 0.55  0.43  0.036 0.273 0.451 0.653 0.439 0.623 0.497 0.56
      0.728 0.452 0.091 0.505 0.788 0.219 0.554 0.748 0.958 0.88  0.945 0.755
      0.444 0.553 0.258 0.562 0.752 0.7   0.524 0.555 0.557 0.547 0.28  0.179
      0.843 0.331 0.713 0.225 0.156 0.216 0.943 0.228 0.437 0.425 0.61  0.497
      0.325 0.517 0.51  0.573 0.683 0.448 0.936 0.986 0.725 0.371 0.984 0.674
      0.528 0.781 0.601 0.744 0.998 0.512 0.115 0.808 0.713 0.632 0.426 0.641
      0.25  0.408 0.875 0.937 0.936 0.785 0.08  0.205 0.573 0.168 0.871 0.791
      0.984 0.071 0.478 0.303 0.527 0.048 0.874 0.626 0.242 0.651 0.736 0.863
      0.838 0.906 0.058 0.979]]
    RECEIVED RESPONSE:
    meta {
      puid: "kir32rkd07l461qt20k5aia2ip"
      requestPath {
        key: "classifier"
        value: "deep-mnist:0.1"
      }
    }
    data {
      names: "class:0"
      names: "class:1"
      names: "class:2"
      names: "class:3"
      names: "class:4"
      names: "class:5"
      names: "class:6"
      names: "class:7"
      names: "class:8"
      names: "class:9"
      ndarray {
        values {
          list_value {
            values {
              number_value: 0.004758698865771294
            }
            values {
              number_value: 2.673733900948605e-09
            }
            values {
              number_value: 0.4710583984851837
            }
            values {
              number_value: 0.145528644323349
            }
            values {
              number_value: 1.7886424785729105e-08
            }
            values {
              number_value: 0.37631428241729736
            }
            values {
              number_value: 2.9651535442098975e-05
            }
            values {
              number_value: 0.0002321827778359875
            }
            values {
              number_value: 0.0020594694651663303
            }
            values {
              number_value: 1.8733135220827535e-05
            }
          }
        }
      }
    }
    
    



```python
!minikube delete
```

    🔥  Deleting "minikube" from virtualbox ...
    💔  The "minikube" cluster has been deleted.



```python

```
