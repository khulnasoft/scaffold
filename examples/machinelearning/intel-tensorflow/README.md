# Tensorflow

[TensorFlow*](https://github.com/tensorflow/tensorflow) is a highly popular machine learning framework
that requires efficient utilization of computational resources. In order to take full advantage of Intel® Xeon® platforms,
the TensorFlow framework has been optimized using the Intel® Math Kernel Library for Deep Neural Networks (Intel® MKL-DNN).

Provided here is all the setup required for Singularity, based on the following Intel Optimized Tensorflow Dockerfiles:
[Dockerfile.mkl](https://github.com/tensorflow/tensorflow/blob/master/tensorflow/tools/docker/Dockerfile.mkl) and [Dockerfile.devel-mkl](https://github.com/tensorflow/tensorflow/blob/master/tensorflow/tools/docker/Dockerfile.devel-mkl)


### Prerequisites
* Singularity (>=3.0.0) - Install from [here](https://github.com/khulnasoft/scaffold/blob/master/INSTALL.md)
* Debootstrap module - Install from [here](https://www.sylabs.io/guides/3.0/user-guide/appendix.html?highlight=debootstrap#id14)

**NOTE:** All definition files where built and tested on`Ubuntu 16.04`.

## Singularity.mkl.def

* This is a light-weight definition file for data scientists. To build this container, download a `python2.7` whl named `intel_tensorflow-1.12.0-cp27-cp27mu-manylinux1_x86_64.whl ` from this [URL](https://pypi.org/project/intel-tensorflow/1.12.0/#files)
and save in the same directory as this README file. (Please use different whl if required, make sure to update all corresponding references)

* Update the `TF_WHL` value in definition file with downloaded wheel name (if different) in `%post` section,
and then read comments in  the `%files` section.

### Build
* To build immutable containers (recommended in production environments):
```
    sudo singularity build <mysingularity_name>.sif Singularity.mkl.def
```
* To build development containers in which you can install additional packages and modify the container.
Please make sure to update those changes in definition file too.
```
    sudo singularity build --sandbox <mysingularity_name>/ Singularity.mkl.def
```

### Shell and Exec
Sample commands for development containers; the same commands apply to immutable containers
without `--writable` option.

`--writable`: option must be provided to install packages.
`--contain`: keeps container’s environment contained, meaning no sharing of host environment.

* Shell
```
    sudo singularity shell --contain --writable <mysingularity_name>/
```

* Exec, to run any arbitary commands from host to inside container
```
    sudo singularity exec --contain <mysingularity_name>/ env
```

### Run a instance
* Runs singularity container in background and allows to access notebook - This will run `%starscript` section commands
```
    sudo singularity instance start <mysingularity_name>/ demotest
```
* List running instances
```
    sudo singularity instance list
```
* To access notebook, you need token. Exec into one of listed instance
```
    sudo singularity exec instance://demotest jupyter notebook list
```
If you are running singularity instance in remote machine.
Execute below command to establish tunnel to forward traffic from remote machine to localhost
Ex: `ssh -L 8888:localhost:8888 <remote_machine_name>`

Access the given url from your browser and enter the access token
Ex: `localhost:8888`

## Singularity.devel-mkl.def
This definition file includes all development tools to build Tensorflow from scratch with MKL configuration.

### Build
* To build immutable containers (recommended in production environments):
```
    sudo singularity build <mysingularity_name>.sif Singularity.devel-mkl.def
```
* To build development containers in which you can install additional packages and modify the container.
Please make sure to update those changes in definition file too.
```
    sudo singularity build --sandbox <mysingularity_name>/ Singularity.devel-mkl.def
```

To Shell and Exec follow the same commands as above. This container does not host any notebooks. This is purely meant for development purpose.

### Run
Executes commands in `%runscript` section
```
sudo singularity run --contain <mysingularity_name>/
```