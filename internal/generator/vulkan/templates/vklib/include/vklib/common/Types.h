#pragma once

/**
 * @file Types.h
 * @brief Tipos e structs compartilhados da biblioteca VulkanLib
 */

#include <vulkan/vulkan.h>
#include <glm/glm.hpp>
#include <cstdint>

namespace vklib {

// ═════════════════════════════════════════════════════════════════════════════
// Tipos de Vertex
// ═════════════════════════════════════════════════════════════════════════════

/**
 * @brief Vértice padrão com posição, normal e coordenada UV
 */
struct Vertex {
    glm::vec3 position;
    glm::vec3 normal;
    glm::vec2 uv;

    // Binding descriptions para Vulkan pipeline
    static VkVertexInputBindingDescription GetBindingDescription();
    static std::array<VkVertexInputAttributeDescription, 3> GetAttributeDescriptions();
};

// ═════════════════════════════════════════════════════════════════════════════
// Estruturas de Uniform Buffer
// ═════════════════════════════════════════════════════════════════════════════

/**
 * @brief Dados de transformação por frame
 */
struct TransformUBO {
    glm::mat4 model;
    glm::mat4 view;
    glm::mat4 projection;
};

/**
 * @brief Dados de câmera
 */
struct CameraUBO {
    glm::vec3 position;
    float     padding0;
    glm::vec3 direction;
    float     padding1;
};

// ═════════════════════════════════════════════════════════════════════════════
// Enums
// ═════════════════════════════════════════════════════════════════════════════

enum class BufferType {
    Vertex,
    Index,
    Uniform,
    Staging,
};

enum class ImageFormat {
    RGBA8,
    Depth32F,
};

}  // namespace vklib
