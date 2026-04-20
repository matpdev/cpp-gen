#pragma once

/**
 * @file DescriptorWriter.h
 * @brief Utilitário para escrever dados em descriptor sets
 */

#include <vulkan/vulkan.h>
#include <vector>

namespace vklib {

/**
 * @brief Utilitário fluent para escrever em descriptor sets
 *
 * Exemplo:
 * ```cpp
 * DescriptorWriter()
 *     .WriteBuffer(0, uniformBuffer.GetBuffer(), sizeof(TransformUBO))
 *     .WriteImage(1, textureSampler, textureImageView, VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL)
 *     .UpdateSet(device, descriptorSet);
 * ```
 */
class DescriptorWriter {
public:
    DescriptorWriter() = default;

    /**
     * @brief Escreve um buffer a um binding
     * @param binding Índice do binding
     * @param buffer VkBuffer
     * @param size Tamanho do buffer
     * @param offset Offset dentro do buffer
     */
    DescriptorWriter& WriteBuffer(uint32_t binding, VkBuffer buffer,
                                   VkDeviceSize size, VkDeviceSize offset = 0);

    /**
     * @brief Escreve uma imagem/sampler a um binding
     * @param binding Índice do binding
     * @param sampler VkSampler
     * @param imageView VkImageView
     * @param layout Layout da imagem
     */
    DescriptorWriter& WriteImage(uint32_t binding, VkSampler sampler,
                                  VkImageView imageView, VkImageLayout layout);

    /**
     * @brief Aplica as escritas ao descriptor set
     * @param device Dispositivo Vulkan
     * @param set VkDescriptorSet
     */
    void UpdateSet(VkDevice device, VkDescriptorSet set);

    /**
     * @brief Limpa os buffers de escrita
     */
    void Clear();

private:
    std::vector<VkWriteDescriptorSet> m_writes;
    std::vector<VkDescriptorBufferInfo> m_bufferInfos;
    std::vector<VkDescriptorImageInfo> m_imageInfos;
};

}  // namespace vklib
