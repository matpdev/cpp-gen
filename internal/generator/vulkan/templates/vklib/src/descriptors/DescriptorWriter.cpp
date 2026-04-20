#include <vklib/common/Defines.h>
#include <vklib/descriptors/DescriptorWriter.h>

#include <cassert>

namespace vklib {
DescriptorWriter& DescriptorWriter::WriteBuffer(uint32_t binding,
                                                VkBuffer buffer,
                                                VkDeviceSize size,
                                                VkDeviceSize offset)
{
    assert(buffer != VK_NULL_HANDLE && "Buffer deve ser válido");

    // Criar info do buffer
    VkDescriptorBufferInfo bufferInfo{};
    bufferInfo.buffer = buffer;
    bufferInfo.offset = offset;
    bufferInfo.range = size;

    // Armazenar para depois usar em UpdateSet
    m_bufferInfos.push_back(bufferInfo);

    // Criar write
    VkWriteDescriptorSet write{};
    write.sType = VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET;
    write.dstBinding = binding;
    write.dstArrayElement = 0;
    write.descriptorCount = 1;
    write.descriptorType = VK_DESCRIPTOR_TYPE_UNIFORM_BUFFER;
    write.pBufferInfo = &m_bufferInfos.back();  // Apontar para a info que acabamos de adicionar

    m_writes.push_back(write);

    return *this;
}

DescriptorWriter& DescriptorWriter::WriteImage(uint32_t binding,
                                               VkSampler sampler,
                                               VkImageView imageView,
                                               VkImageLayout layout)
{
    assert(sampler != VK_NULL_HANDLE && "Sampler deve ser válido");
    assert(imageView != VK_NULL_HANDLE && "ImageView deve ser válido");

    // Criar info da imagem
    VkDescriptorImageInfo imageInfo{};
    imageInfo.sampler = sampler;
    imageInfo.imageView = imageView;
    imageInfo.imageLayout = layout;

    // Armazenar para depois usar em UpdateSet
    m_imageInfos.push_back(imageInfo);

    // Criar write
    VkWriteDescriptorSet write{};
    write.sType = VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET;
    write.dstBinding = binding;
    write.dstArrayElement = 0;
    write.descriptorCount = 1;
    write.descriptorType = VK_DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER;
    write.pImageInfo = &m_imageInfos.back();  // Apontar para a info que acabamos de adicionar

    m_writes.push_back(write);

    return *this;
}

void DescriptorWriter::UpdateSet(VkDevice device, VkDescriptorSet set)
{
    assert(device != VK_NULL_HANDLE && "Device deve ser válido");
    assert(set != VK_NULL_HANDLE && "DescriptorSet deve ser válido");
    assert(!m_writes.empty() && "Nenhum write foi adicionado");

    // Setar o descriptor set para todos os writes
    for (auto& write : m_writes) {
        write.dstSet = set;
    }

    // Atualizar descriptors
    vkUpdateDescriptorSets(
        device, static_cast<uint32_t>(m_writes.size()), m_writes.data(), 0, nullptr);

    VK_LOG_INFO("Descriptor Set atualizado com " << m_writes.size() << " writes");
}

void DescriptorWriter::Clear()
{
    m_writes.clear();
    m_bufferInfos.clear();
    m_imageInfos.clear();
}
}  // namespace vklib
