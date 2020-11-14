
find_package(Protobuf REQUIRED)
find_package(PythonLibs REQUIRED)

add_subdirectory(../pybind11 ${CMAKE_CURRENT_BINARY_DIR}/pybind_build})

if(SELDON_OPT_BUILD_PROTO)
    execute_process(COMMAND make -C ${PROJECT_SOURCE_DIR}/../
 proto-build)
endif()
if(SELDON_OPT_CLONE_PYBIND11)
    execute_process(COMMAND make -C ${PROJECT_SOURCE_DIR}/../
 clone-pybind)
endif()

file(GLOB SELDON_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/*.cpp")

file(GLOB SELDON_PROTO_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/proto/*.cc")

if(SELDON_OPT_MAKE_SHARED)
    add_library(
        seldon SHARED
        ${SELDON_SRC} ${SELDON_PROTO_SRC})
else()
    add_library(
        seldon STATIC
        ${SELDON_SRC} ${SELDON_PROTO_SRC})
endif()

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

if(MSVC)
    target_link_libraries(
        seldon PRIVATE 
        pybind11::windows_extras)
endif()

add_library(seldon::seldon ALIAS seldon)


