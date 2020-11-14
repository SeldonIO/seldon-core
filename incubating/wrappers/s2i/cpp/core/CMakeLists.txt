cmake_minimum_required(VERSION 3.4.1)
project(seldon VERSION 0.0.1)

option(SELDON_OPT_BUILD_TESTS "Enable if you want to build tests" 0)
option(SELDON_OPT_MAKE_SHARED "Builds static library by default but can build as shared if set to true" 0)
option(SELDON_OPT_BUILD_PROTO "Builds proto cpp files using protoc" 0)
option(SELDON_OPT_INSTALL "Whether to set up rules to install" 0)
option(SELDON_OPT_CLONE_PYBIND11 "Whether to clone the pybind11 repository for development" 1)

set(CMAKE_CXX_STANDARD 14)

# Allow scripts to call main kompute Makefile
function(seldon_make SELDON_MAKE_TARGET)
    add_custom_target(${SELDON_MAKE_TARGET}
        COMMAND make -C ${PROJECT_SOURCE_DIR}/../ ${SELDON_MAKE_TARGET})
endfunction()

add_subdirectory(src)

if(SELDON_OPT_BUILD_TESTS)
    add_subdirectory(test)
endif()

