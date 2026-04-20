#pragma once

/**
 * @file FrameData.h
 * @brief Encapsulação de dados por frame (CommandBuffer, Semáforo de Aquisição, Fence)
 *
 * NOTA: O semáforo de render (sinalizado pelo submit, aguardado pelo present)
 * é gerenciado POR IMAGEM da swapchain em EngineContext::m_renderSemaphores,
 * evitando a reutilização prematura de semáforos (VUID-vkQueueSubmit-00067).
 */

#include <vulkan/vulkan.h>

#include <utility>
#include <vector>

namespace vklib {

class FrameData
{
public:
    FrameData() = default;
    ~FrameData();

    FrameData(const FrameData&) = delete;
    FrameData& operator=(const FrameData&) = delete;

    FrameData(FrameData&& other) noexcept;
    FrameData& operator=(FrameData&& other) noexcept;

    void Initialize(VkDevice device, VkCommandPool commandPool);
    void Destroy(VkDevice device, VkCommandPool commandPool);

    VkCommandBuffer GetCommandBuffer() const
    {
        return m_commandBuffer;
    }
    VkSemaphore GetAcquireSemaphore() const
    {
        return m_acquireSemaphore;
    }
    VkFence GetRenderFence() const
    {
        return m_renderFence;
    }

private:
    VkCommandBuffer m_commandBuffer = VK_NULL_HANDLE;
    VkSemaphore m_acquireSemaphore = VK_NULL_HANDLE;  // sinalizado por vkAcquireNextImageKHR
    VkFence m_renderFence = VK_NULL_HANDLE;
};

}  // namespace vklib
