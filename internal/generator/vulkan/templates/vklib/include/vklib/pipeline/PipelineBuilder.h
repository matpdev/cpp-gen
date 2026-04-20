#pragma once

/**
 * @file PipelineBuilder.h
 * @brief Builder Pattern para construir pipelines Vulkan
 */

#include <vulkan/vulkan.h>
#include "Pipeline.h"
#include <vector>
#include <string_view>
#include <memory>

namespace vklib {

/**
 * @brief Builder para criar pipelines gráficos com fluent API
 *
 * Exemplo:
 * ```cpp
 * auto pipeline = PipelineBuilder()
 *     .SetDevice(device)
 *     .SetRenderPass(renderPass)
 *     .SetViewport(0, 0, width, height)
 *     .AddShader(VK_SHADER_STAGE_VERTEX_BIT, vertShaderModule, "main")
 *     .AddShader(VK_SHADER_STAGE_FRAGMENT_BIT, fragShaderModule, "main")
 *     .Build();
 * ```
 */
class PipelineBuilder {
public:
    PipelineBuilder() = default;

    // ────────────────────────────────────────────────────────────────────────
    // Configurações
    // ────────────────────────────────────────────────────────────────────────

    PipelineBuilder& SetDevice(VkDevice device);
    PipelineBuilder& SetRenderPass(VkRenderPass renderPass, uint32_t subpass = 0);
    PipelineBuilder& SetViewport(float x, float y, float width, float height);
    PipelineBuilder& SetScissor(int32_t x, int32_t y, uint32_t width, uint32_t height);

    /**
     * @brief Adiciona um shader stage
     * @param stage VK_SHADER_STAGE_VERTEX_BIT, VK_SHADER_STAGE_FRAGMENT_BIT, etc.
     * @param module Módulo shader (VkShaderModule)
     * @param entryPoint Nome da função (geralmente "main")
     */
    PipelineBuilder& AddShader(VkShaderStageFlagBits stage, VkShaderModule module, std::string_view entryPoint);

    /**
     * @brief Define o vertex input binding e attributes
     */
    PipelineBuilder& SetVertexInput(const std::vector<VkVertexInputBindingDescription>& bindings,
                                     const std::vector<VkVertexInputAttributeDescription>& attributes);

    /**
     * @brief Define o tipo de topologia (triangles, lines, points, etc.)
     */
    PipelineBuilder& SetTopology(VkPrimitiveTopology topology);

    /**
     * @brief Define configurações de rasterização
     */
    PipelineBuilder& SetRasterization(VkPolygonMode polygonMode = VK_POLYGON_MODE_FILL,
                                       VkCullModeFlags cullMode = VK_CULL_MODE_BACK_BIT,
                                       VkFrontFace frontFace = VK_FRONT_FACE_COUNTER_CLOCKWISE);

    /**
     * @brief Define configurações de multisample
     */
    PipelineBuilder& SetMultisample(VkSampleCountFlagBits sampleCount = VK_SAMPLE_COUNT_1_BIT);

    /**
     * @brief Define configurações de color blending
     */
    PipelineBuilder& SetColorBlending(VkBool32 blendEnable = VK_FALSE);

    /**
     * @brief Define o layout de descritores
     */
    PipelineBuilder& SetDescriptorSetLayout(VkDescriptorSetLayout layout);

    // ────────────────────────────────────────────────────────────────────────
    // Build
    // ────────────────────────────────────────────────────────────────────────

    /**
     * @brief Constrói a pipeline
     * @return Pipeline configurada
     */
    std::unique_ptr<Pipeline> Build();

private:
    VkDevice m_device = VK_NULL_HANDLE;
    VkRenderPass m_renderPass = VK_NULL_HANDLE;
    uint32_t m_subpass = 0;

    std::vector<VkPipelineShaderStageCreateInfo> m_shaderStages;
    std::vector<const char*> m_shaderEntryPoints;

    VkViewport m_viewport{};
    VkRect2D m_scissor{};

    VkVertexInputBindingDescription m_bindingDescription{};
    std::vector<VkVertexInputAttributeDescription> m_attributeDescriptions;

    VkPrimitiveTopology m_topology = VK_PRIMITIVE_TOPOLOGY_TRIANGLE_LIST;

    VkPipelineRasterizationStateCreateInfo m_rasterization{};
    VkPipelineMultisampleStateCreateInfo m_multisample{};
    VkPipelineColorBlendAttachmentState m_colorBlendAttachment{};
    VkDescriptorSetLayout m_descriptorSetLayout = VK_NULL_HANDLE;
};

}  // namespace vklib
