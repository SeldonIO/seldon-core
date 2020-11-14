
include(FindProtobuf)
find_package(Protobuf REQUIRED)

find_package(PythonLibs REQUIRED)
#find_package(Python COMPONENTS Interpreter Development REQUIRED)

add_subdirectory(../../pybind11 ${CMAKE_CURRENT_BINARY_DIR}/pybind_build})

file(GLOB SELDON_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/*.cpp")

file(GLOB SELDON_PROTO_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/proto/*.cc")

#pybind11_add_module(
#    seldon
#    SHARED ${SELDON_SRC} ${SELDON_PROTO_SRC})

# Build static library as well
add_library(
    seldon SHARED
    ${SELDON_SRC} ${SELDON_PROTO_SRC})

target_include_directories(
    seldon PUBLIC
    $<INSTALL_INTERFACE:include>
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
)

target_include_directories(
    seldon PUBLIC 
    ${PYTHON_INCLUDE_DIRS})

target_link_libraries(
    seldon PRIVATE
    pybind11::module
    pybind11::lto
    ${PROTOBUF_LIBRARIES})

pybind11_extension(seldon)
pybind11_strip(seldon)

if(MSVC)
    target_link_libraries(
        seldon PRIVATE 
        pybind11::windows_extras)
endif()

add_library(seldon::seldon ALIAS seldon)


