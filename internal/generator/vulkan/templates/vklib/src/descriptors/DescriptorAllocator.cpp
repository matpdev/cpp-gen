#include <vklib/common/Defines.h>
#include <vklib/descriptors/DescriptorAllocator.h>
#include <vulkan/vulkan_core.h>

#include <cassert>
#include <cstdint>
#include <vector>

namespace vklib {
DescriptorAllocator::~DescriptorAllocator()
{
    if (m_pool != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
        vkDestroyDescriptorPool(m_device, m_pool, nullptr);
        m_pool = VK_NULL_HANDLE;
    }
    m_device = VK_NULL_HANDLE;
}

void DescriptorAllocator::Initialize(VkDevice device, uint32_t maxSets)
{
    assert(device != VK_NULL_HANDLE && "Device deve ser válido");
    assert(maxSets > 0 && "maxSets deve ser > 0");

    m_device = device;

    std::vector<VkDescriptorPoolSize> poolSizes;

    // Uniform Buffers
    VkDescriptorPoolSize uniformBufferPoolSize{};
    uniformBufferPoolSize.type = VK_DESCRIPTOR_TYPE_UNIFORM_BUFFER;
    uniformBufferPoolSize.descriptorCount = maxSets * 2;  // 2 por set
    poolSizes.push_back(uniformBufferPoolSize);

    // Storage Buffers (opcional)
    VkDescriptorPoolSize storageBufferPoolSize{};
    storageBufferPoolSize.type = VK_DESCRIPTOR_TYPE_STORAGE_BUFFER;
    storageBufferPoolSize.descriptorCount = maxSets * 2;
    poolSizes.push_back(storageBufferPoolSize);

    // Combined Image Samplers
    VkDescriptorPoolSize imageSamplerPoolSize{};
    imageSamplerPoolSize.type = VK_DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER;
    imageSamplerPoolSize.descriptorCount = maxSets * 4;  // 4 texturas por set
    poolSizes.push_back(imageSamplerPoolSize);

    // Storage Images (opcional)
    VkDescriptorPoolSize storageImagePoolSize{};
    storageImagePoolSize.type = VK_DESCRIPTOR_TYPE_STORAGE_IMAGE;
    storageImagePoolSize.descriptorCount = maxSets * 2;
    poolSizes.push_back(storageImagePoolSize);

    VkDescriptorPoolCreateInfo poolInfo{};
    poolInfo.sType = VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO;
    poolInfo.flags =
        VK_DESCRIPTOR_POOL_CREATE_FREE_DESCRIPTOR_SET_BIT;  // Permite liberação individual
    poolInfo.maxSets = maxSets;
    poolInfo.poolSizeCount = static_cast<uint32_t>(poolSizes.size());
    poolInfo.pPoolSizes = poolSizes.data();

    VK_CHECK(vkCreateDescriptorPool(m_device, &poolInfo, nullptr, &m_pool));

    VK_LOG_INFO("Descriptor Pool criada: " << maxSets << " sets máximo");
}

VkDescriptorSet DescriptorAllocator::Allocate(VkDescriptorSetLayout layout)
{
    assert(m_device != VK_NULL_HANDLE && "DescriptorAllocator não foi inicializado");
    assert(m_pool != VK_NULL_HANDLE && "Descriptor Pool não foi criada");
    assert(layout != VK_NULL_HANDLE && "Layout deve ser válido");

    VkDescriptorSetAllocateInfo allocInfo{};
    allocInfo.sType = VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO;
    allocInfo.descriptorPool = m_pool;
    allocInfo.descriptorSetCount = 1;
    allocInfo.pSetLayouts = &layout;

    VkDescriptorSet set = VK_NULL_HANDLE;
    VkResult result = vkAllocateDescriptorSets(m_device, &allocInfo, &set);

    if (result != VK_SUCCESS) {
        VK_LOG_ERROR("Falha ao alocar Descriptor Set: " << static_cast<int>(result));
        return VK_NULL_HANDLE;
    }

    return set;
}

void DescriptorAllocator::Reset()
{
    if (m_pool != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
        VK_CHECK(vkResetDescriptorPool(m_device, m_pool, 0));
        VK_LOG_INFO("Descriptor Pool resetada");
    }
}

}  // namespace vklib
