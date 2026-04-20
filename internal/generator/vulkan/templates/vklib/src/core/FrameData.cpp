#include <vklib/common/Defines.h>
#include <vklib/core/FrameData.h>
#include <vulkan/vulkan_core.h>

namespace vklib {

FrameData::~FrameData() {}

FrameData::FrameData(FrameData&& other) noexcept
    : m_commandBuffer(std::exchange(other.m_commandBuffer, VK_NULL_HANDLE)),
      m_acquireSemaphore(std::exchange(other.m_acquireSemaphore, VK_NULL_HANDLE)),
      m_renderFence(std::exchange(other.m_renderFence, VK_NULL_HANDLE))
{}

FrameData& FrameData::operator=(FrameData&& other) noexcept
{
    if (this != &other) {
        m_commandBuffer = std::exchange(other.m_commandBuffer, VK_NULL_HANDLE);
        m_acquireSemaphore = std::exchange(other.m_acquireSemaphore, VK_NULL_HANDLE);
        m_renderFence = std::exchange(other.m_renderFence, VK_NULL_HANDLE);
    }
    return *this;
}

void FrameData::Initialize(VkDevice device, VkCommandPool commandPool)
{
    VkCommandBufferAllocateInfo cmdAllocInfo{};
    cmdAllocInfo.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO;
    cmdAllocInfo.commandPool = commandPool;
    cmdAllocInfo.level = VK_COMMAND_BUFFER_LEVEL_PRIMARY;
    cmdAllocInfo.commandBufferCount = 1;
    VK_CHECK(vkAllocateCommandBuffers(device, &cmdAllocInfo, &m_commandBuffer));

    // Semáforo de aquisição: sinalizado quando a imagem está disponível para render
    VkSemaphoreCreateInfo semaphoreInfo{};
    semaphoreInfo.sType = VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO;
    VK_CHECK(vkCreateSemaphore(device, &semaphoreInfo, nullptr, &m_acquireSemaphore));

    // Fence começa sinalizada para não travar no primeiro frame
    VkFenceCreateInfo fenceInfo{};
    fenceInfo.sType = VK_STRUCTURE_TYPE_FENCE_CREATE_INFO;
    fenceInfo.flags = VK_FENCE_CREATE_SIGNALED_BIT;
    VK_CHECK(vkCreateFence(device, &fenceInfo, nullptr, &m_renderFence));
}

void FrameData::Destroy(VkDevice device, VkCommandPool commandPool)
{
    if (device == VK_NULL_HANDLE)
        return;

    if (m_commandBuffer != VK_NULL_HANDLE) {
        vkFreeCommandBuffers(device, commandPool, 1, &m_commandBuffer);
        m_commandBuffer = VK_NULL_HANDLE;
    }
    if (m_acquireSemaphore != VK_NULL_HANDLE) {
        vkDestroySemaphore(device, m_acquireSemaphore, nullptr);
        m_acquireSemaphore = VK_NULL_HANDLE;
    }
    if (m_renderFence != VK_NULL_HANDLE) {
        vkDestroyFence(device, m_renderFence, nullptr);
        m_renderFence = VK_NULL_HANDLE;
    }
}

}  // namespace vklib
