#pragma once

/**
 * @file Pipeline.h
 * @brief Abstração para VkPipeline (graphics pipeline)
 */

#include <vulkan/vulkan.h>

#include <memory>
#include <utility>
#include <vector>

namespace vklib {

/**
 * @brief Pipeline gráfico Vulkan
 *
 * Encapsula VkPipeline e VkPipelineLayout.
 */
class Pipeline
{
public:
    Pipeline() = default;
    ~Pipeline();

    // Não permitir cópia
    Pipeline(const Pipeline&) = delete;
    Pipeline& operator=(const Pipeline&) = delete;

    // Mover é permitido
    Pipeline(Pipeline&& other) noexcept;
    Pipeline& operator=(Pipeline&& other) noexcept;

    // ────────────────────────────────────────────────────────────────────────
    // Getters (Escape Hatches)
    // ────────────────────────────────────────────────────────────────────────

    VkPipeline GetPipeline() const
    {
        return m_pipeline;
    }
    VkPipelineLayout GetLayout() const
    {
        return m_layout;
    }

private:
    friend class PipelineBuilder;

    VkPipeline m_pipeline = VK_NULL_HANDLE;
    VkPipelineLayout m_layout = VK_NULL_HANDLE;
    VkDevice m_device = VK_NULL_HANDLE;
};

}  // namespace vklib
