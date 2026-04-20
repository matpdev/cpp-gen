#include <vklib/common/Defines.h>
#include <vklib/pipeline/PipelineBuilder.h>

#include <cassert>
#include <cstring>
#include <vector>

namespace vklib {
PipelineBuilder& PipelineBuilder::SetDevice(VkDevice device)
{
    m_device = device;
    return *this;
}

PipelineBuilder& PipelineBuilder::SetRenderPass(VkRenderPass renderPass, uint32_t subpass)
{
    m_renderPass = renderPass;
    m_subpass = subpass;
    return *this;
}

PipelineBuilder& PipelineBuilder::SetViewport(float x, float y, float width, float height)
{
    m_viewport.x = x;
    m_viewport.y = y;
    m_viewport.width = width;
    m_viewport.height = height;
    m_viewport.minDepth = 0.0f;
    m_viewport.maxDepth = 1.0f;

    // Scissor padrão: mesma área do viewport
    m_scissor.offset = {static_cast<int32_t>(x), static_cast<int32_t>(y)};
    m_scissor.extent = {static_cast<uint32_t>(width), static_cast<uint32_t>(height)};

    return *this;
}

PipelineBuilder& PipelineBuilder::SetScissor(int32_t x, int32_t y, uint32_t width, uint32_t height)
{
    m_scissor.offset = {x, y};
    m_scissor.extent = {width, height};
    return *this;
}

PipelineBuilder& PipelineBuilder::AddShader(VkShaderStageFlagBits stage,
                                            VkShaderModule module,
                                            std::string_view entryPoint)
{
    VkPipelineShaderStageCreateInfo shaderStage{};
    shaderStage.sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO;
    shaderStage.stage = stage;
    shaderStage.module = module;
    // Armazenar o entryPoint string separadamente
    shaderStage.pName = nullptr;  // Será preenchido no Build()

    m_shaderStages.push_back(shaderStage);
    m_shaderEntryPoints.push_back(entryPoint.data());

    return *this;
}

PipelineBuilder& PipelineBuilder::SetVertexInput(
    const std::vector<VkVertexInputBindingDescription>& bindings,
    const std::vector<VkVertexInputAttributeDescription>& attributes)
{

    if (!bindings.empty()) {
        m_bindingDescription = bindings[0];
    }
    m_attributeDescriptions = attributes;

    return *this;
}

PipelineBuilder& PipelineBuilder::SetTopology(VkPrimitiveTopology topology)
{
    m_topology = topology;
    return *this;
}

PipelineBuilder& PipelineBuilder::SetRasterization(VkPolygonMode polygonMode,
                                                   VkCullModeFlags cullMode,
                                                   VkFrontFace frontFace)
{
    m_rasterization.sType = VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO;
    m_rasterization.depthClampEnable = VK_FALSE;
    m_rasterization.rasterizerDiscardEnable = VK_FALSE;
    m_rasterization.polygonMode = polygonMode;
    m_rasterization.cullMode = cullMode;
    m_rasterization.frontFace = frontFace;
    m_rasterization.depthBiasEnable = VK_FALSE;
    m_rasterization.lineWidth = 1.0f;

    return *this;
}

PipelineBuilder& PipelineBuilder::SetMultisample(VkSampleCountFlagBits sampleCount)
{
    m_multisample.sType = VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO;
    m_multisample.rasterizationSamples = sampleCount;
    m_multisample.sampleShadingEnable = VK_FALSE;
    m_multisample.minSampleShading = 1.0f;
    m_multisample.pSampleMask = nullptr;
    m_multisample.alphaToCoverageEnable = VK_FALSE;
    m_multisample.alphaToOneEnable = VK_FALSE;

    return *this;
}

PipelineBuilder& PipelineBuilder::SetColorBlending(VkBool32 blendEnable)
{
    m_colorBlendAttachment.colorWriteMask = VK_COLOR_COMPONENT_R_BIT | VK_COLOR_COMPONENT_G_BIT |
                                            VK_COLOR_COMPONENT_B_BIT | VK_COLOR_COMPONENT_A_BIT;
    m_colorBlendAttachment.blendEnable = blendEnable;

    if (blendEnable) {
        m_colorBlendAttachment.srcColorBlendFactor = VK_BLEND_FACTOR_SRC_ALPHA;
        m_colorBlendAttachment.dstColorBlendFactor = VK_BLEND_FACTOR_ONE_MINUS_SRC_ALPHA;
        m_colorBlendAttachment.colorBlendOp = VK_BLEND_OP_ADD;
        m_colorBlendAttachment.srcAlphaBlendFactor = VK_BLEND_FACTOR_ONE;
        m_colorBlendAttachment.dstAlphaBlendFactor = VK_BLEND_FACTOR_ZERO;
        m_colorBlendAttachment.alphaBlendOp = VK_BLEND_OP_ADD;
    }

    return *this;
}

PipelineBuilder& PipelineBuilder::SetDescriptorSetLayout(VkDescriptorSetLayout layout)
{
    m_descriptorSetLayout = layout;
    return *this;
}

std::unique_ptr<Pipeline> PipelineBuilder::Build()
{
    assert(m_device != VK_NULL_HANDLE && "Device deve ser setado");
    assert(m_renderPass != VK_NULL_HANDLE && "RenderPass deve ser setado");
    assert(!m_shaderStages.empty() && "Pelo menos um shader deve ser adicionado");

    auto pipeline = std::make_unique<Pipeline>();
    pipeline->m_device = m_device;

    for (size_t i = 0; i < m_shaderStages.size(); ++i) {
        m_shaderStages[i].pName = m_shaderEntryPoints[i];
    }

    VkPipelineVertexInputStateCreateInfo vertexInputInfo{};
    vertexInputInfo.sType = VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO;

    if (!m_attributeDescriptions.empty()) {
        vertexInputInfo.vertexBindingDescriptionCount = 1;
        vertexInputInfo.pVertexBindingDescriptions = &m_bindingDescription;
        vertexInputInfo.vertexAttributeDescriptionCount =
            static_cast<uint32_t>(m_attributeDescriptions.size());
        vertexInputInfo.pVertexAttributeDescriptions = m_attributeDescriptions.data();
    }

    VkPipelineInputAssemblyStateCreateInfo inputAssembly{};
    inputAssembly.sType = VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO;
    inputAssembly.topology = m_topology;
    inputAssembly.primitiveRestartEnable = VK_FALSE;

    VkPipelineViewportStateCreateInfo viewportState{};
    viewportState.sType = VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO;
    viewportState.viewportCount = 1;
    viewportState.pViewports = &m_viewport;
    viewportState.scissorCount = 1;
    viewportState.pScissors = &m_scissor;

    VkPipelineColorBlendStateCreateInfo colorBlending{};
    colorBlending.sType = VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO;
    colorBlending.logicOpEnable = VK_FALSE;
    colorBlending.logicOp = VK_LOGIC_OP_COPY;
    colorBlending.attachmentCount = 1;
    colorBlending.pAttachments = &m_colorBlendAttachment;
    colorBlending.blendConstants[0] = 0.0f;
    colorBlending.blendConstants[1] = 0.0f;
    colorBlending.blendConstants[2] = 0.0f;
    colorBlending.blendConstants[3] = 0.0f;

    VkPipelineLayoutCreateInfo layoutInfo{};
    layoutInfo.sType = VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO;

    if (m_descriptorSetLayout != VK_NULL_HANDLE) {
        layoutInfo.setLayoutCount = 1;
        layoutInfo.pSetLayouts = &m_descriptorSetLayout;
    }

    VK_CHECK(vkCreatePipelineLayout(m_device, &layoutInfo, nullptr, &pipeline->m_layout));

    if (m_rasterization.sType == 0) {
        m_rasterization.sType = VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO;
        m_rasterization.depthClampEnable = VK_FALSE;
        m_rasterization.rasterizerDiscardEnable = VK_FALSE;
        m_rasterization.polygonMode = VK_POLYGON_MODE_FILL;
        m_rasterization.cullMode = VK_CULL_MODE_BACK_BIT;
        m_rasterization.frontFace = VK_FRONT_FACE_COUNTER_CLOCKWISE;
        m_rasterization.depthBiasEnable = VK_FALSE;
        m_rasterization.lineWidth = 1.0f;
    }

    if (m_multisample.sType == 0) {
        m_multisample.sType = VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO;
        m_multisample.rasterizationSamples = VK_SAMPLE_COUNT_1_BIT;
        m_multisample.sampleShadingEnable = VK_FALSE;
    }

    VkGraphicsPipelineCreateInfo pipelineInfo{};
    pipelineInfo.sType = VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO;
    pipelineInfo.stageCount = static_cast<uint32_t>(m_shaderStages.size());
    pipelineInfo.pStages = m_shaderStages.data();
    pipelineInfo.pVertexInputState = &vertexInputInfo;
    pipelineInfo.pInputAssemblyState = &inputAssembly;
    pipelineInfo.pViewportState = &viewportState;
    pipelineInfo.pRasterizationState = &m_rasterization;
    pipelineInfo.pMultisampleState = &m_multisample;
    pipelineInfo.pColorBlendState = &colorBlending;
    pipelineInfo.layout = pipeline->m_layout;
    pipelineInfo.renderPass = m_renderPass;
    pipelineInfo.subpass = m_subpass;
    pipelineInfo.basePipelineHandle = VK_NULL_HANDLE;
    pipelineInfo.basePipelineIndex = -1;

    VK_CHECK(vkCreateGraphicsPipelines(
        m_device, VK_NULL_HANDLE, 1, &pipelineInfo, nullptr, &pipeline->m_pipeline));

    VK_LOG_INFO("Graphics Pipeline criada com sucesso");

    return pipeline;
}
}  // namespace vklib
