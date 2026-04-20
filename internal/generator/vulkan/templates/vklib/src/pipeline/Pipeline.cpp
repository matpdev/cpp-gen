#include <vklib/common/Defines.h>
#include <vklib/pipeline/Pipeline.h>

namespace vklib {
Pipeline::~Pipeline()
{
    if (m_pipeline != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
        vkDestroyPipeline(m_device, m_pipeline, nullptr);
        m_pipeline = VK_NULL_HANDLE;
    }

    if (m_layout != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
        vkDestroyPipelineLayout(m_device, m_layout, nullptr);
        m_layout = VK_NULL_HANDLE;
    }

    m_device = VK_NULL_HANDLE;
}

Pipeline::Pipeline(Pipeline&& other) noexcept
    : m_pipeline(std::exchange(other.m_pipeline, VK_NULL_HANDLE)),
      m_layout(std::exchange(other.m_layout, VK_NULL_HANDLE)),
      m_device(std::exchange(other.m_device, VK_NULL_HANDLE))
{}

Pipeline& Pipeline::operator=(Pipeline&& other) noexcept
{
    if (this != &other) {
        // Destruir recursos existentes
        if (m_pipeline != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
            vkDestroyPipeline(m_device, m_pipeline, nullptr);
        }
        if (m_layout != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
            vkDestroyPipelineLayout(m_device, m_layout, nullptr);
        }

        // Mover recursos
        m_pipeline = std::exchange(other.m_pipeline, VK_NULL_HANDLE);
        m_layout = std::exchange(other.m_layout, VK_NULL_HANDLE);
        m_device = std::exchange(other.m_device, VK_NULL_HANDLE);
    }
    return *this;
}
}  // namespace vklib
