#pragma once

/**
 * @file EngineContext.h
 * @brief Contexto global da engine, gerencia VkInstance, Device, Swapchain, etc.
 */

#include "FrameData.h"

#define GLFW_INCLUDE_VULKAN
#include "GLFW/glfw3.h"

#include <vulkan/vulkan.h>

#include <memory>
#include <vector>

#include <VkBootstrap.h>

struct GLFWwindow;

namespace vklib {

/**
 * @brief Contexto central da engine Vulkan
 *
 * Gerencia:
 * - VkInstance, VkPhysicalDevice, VkDevice via VkBootstrap
 * - Swapchain e Renderpass
 * - Frame syncing (Double Buffering)
 * - Loop de apresentação básico
 */
class EngineContext
{
public:
    EngineContext() = default;
    ~EngineContext();

    // Não permitir cópia
    EngineContext(const EngineContext&) = delete;
    EngineContext& operator=(const EngineContext&) = delete;

    // ────────────────────────────────────────────────────────────────────────
    // Inicialização
    // ────────────────────────────────────────────────────────────────────────

    /**
     * @brief Inicializa a engine com uma janela GLFW
     * @param windowTitle Título da janela
     * @param width Largura em pixels
     * @param height Altura em pixels
     * @return true se inicialização bem-sucedida, false caso contrário
     */
    bool Initialize(const char* windowTitle, uint32_t width, uint32_t height);

    /**
     * @brief Libera todos os recursos Vulkan e fecha a janela
     */
    void Shutdown();

    // ────────────────────────────────────────────────────────────────────────
    // Loop Principal
    // ────────────────────────────────────────────────────────────────────────

    /**
     * @brief Verifica se a janela deve ser fechada
     * @return true se deve continuar, false se deve sair
     */
    bool IsRunning() const;

    /**
     * @brief Inicia um novo frame
     * @return Referência ao FrameData atual para gravação de comandos
     */
    FrameData& BeginFrame();

    /**
     * @brief Finaliza o frame e submete para renderização
     */
    void EndFrame();

    // ────────────────────────────────────────────────────────────────────────
    // Getters (Escape Hatches)
    // ────────────────────────────────────────────────────────────────────────

    VkDevice GetDevice() const
    {
        return m_device;
    }
    VkPhysicalDevice GetPhysicalDevice() const
    {
        return m_physicalDevice;
    }
    VkQueue GetGraphicsQueue() const
    {
        return m_graphicsQueue;
    }
    uint32_t GetGraphicsQueueFamily() const
    {
        return m_graphicsQueueFamily;
    }
    VkSwapchainKHR GetSwapchain() const
    {
        return m_swapchain;
    }
    VkExtent2D GetSwapchainExtent() const
    {
        return m_swapchainExtent;
    }
    VkFormat GetSwapchainFormat() const
    {
        return m_swapchainFormat;
    }
    VkRenderPass GetRenderPass() const
    {
        return m_renderPass;
    }
    GLFWwindow* GetWindow() const
    {
        return m_window;
    }
    uint32_t GetCurrentFrameIndex() const
    {
        return m_currentFrame;
    }
    VkInstance GetInstance() const
    {
        return m_instance;
    }
    bool IsFrameReady() const
    {
        return m_frameReady;
    }

private:
    // ────────────────────────────────────────────────────────────────────────
    // Helpers privados
    // ────────────────────────────────────────────────────────────────────────

    bool InitializeVulkan();
    bool InitializeSwapchain();
    bool InitializeRenderPass();
    bool InitializeFrameSyncing();

    void RecreateSwapchain();
    void DestroySwapchain();
    void DestroyFrameSyncing();

    // ────────────────────────────────────────────────────────────────────────
    // Membros
    // ────────────────────────────────────────────────────────────────────────

    GLFWwindow* m_window = nullptr;
    uint32_t m_windowWidth = 0;
    uint32_t m_windowHeight = 0;

    // Vulkan Core
    VkInstance m_instance = VK_NULL_HANDLE;
    VkPhysicalDevice m_physicalDevice = VK_NULL_HANDLE;
    VkDevice m_device = VK_NULL_HANDLE;
    VkQueue m_graphicsQueue = VK_NULL_HANDLE;
    uint32_t m_graphicsQueueFamily = 0;

    VkSurfaceKHR m_surface = VK_NULL_HANDLE;

    // Swapchain
    VkSwapchainKHR m_swapchain = VK_NULL_HANDLE;
    VkExtent2D m_swapchainExtent{};
    VkFormat m_swapchainFormat = VK_FORMAT_UNDEFINED;
    std::vector<VkImage> m_swapchainImages;
    std::vector<VkImageView> m_swapchainImageViews;
    std::vector<VkFramebuffer> m_framebuffers;

    // Renderpass
    VkRenderPass m_renderPass = VK_NULL_HANDLE;

    // Command Buffers
    VkCommandPool m_commandPool = VK_NULL_HANDLE;

    // Frame Syncing
    std::vector<FrameData> m_frames;
    std::vector<VkSemaphore> m_renderSemaphores;
    uint32_t m_currentFrame = 0;
    uint32_t m_swapchainImageIndex = 0;
    bool m_frameReady = false;

    // Validação
    VkDebugUtilsMessengerEXT m_debugMessenger = VK_NULL_HANDLE;
};

}  // namespace vklib
