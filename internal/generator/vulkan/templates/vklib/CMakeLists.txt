# =============================================================================
# vklib/CMakeLists.txt — Biblioteca de abstração Vulkan
# =============================================================================
# Este módulo encapsula toda a camada de abstração Vulkan (vklib).
# Ele é compilado como uma biblioteca estática e linkado ao executável
# principal em src/.
# =============================================================================

add_library(vklib STATIC
    src/vma_implementation.cpp

    # Common
    src/common/VkCheck.cpp
    src/common/Types.cpp

    # Core
    src/core/FrameData.cpp
    src/core/EngineContext.cpp

    # Memory
    src/memory/Buffer.cpp
    src/memory/Image.cpp

    # Pipeline
    src/pipeline/Pipeline.cpp
    src/pipeline/PipelineBuilder.cpp
    src/pipeline/ShaderLoader.cpp

    # Descriptors
    src/descriptors/DescriptorAllocator.cpp
    src/descriptors/DescriptorWriter.cpp

    # System
    src/system/SystemInfo.cpp
)

target_include_directories(vklib
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(vklib
    PUBLIC
        Vulkan::Vulkan
        vk-bootstrap::vk-bootstrap
        GPUOpen::VulkanMemoryAllocator
        glfw
        glm::glm
    PRIVATE
        project_warnings
)

set_target_properties(vklib PROPERTIES
    CXX_STANDARD          23
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
)
