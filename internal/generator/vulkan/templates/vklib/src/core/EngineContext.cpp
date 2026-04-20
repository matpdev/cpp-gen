
#include "vklib/common/Defines.h"
#include "vklib/core/EngineContext.h"
#include "vklib/core/FrameData.h"

#include <vulkan/vulkan_core.h>

#include <algorithm>
#include <cstdint>
#include <iostream>

namespace vklib {
// ═════════════════════════════════════════════════════════════════════════════
// Destrutor
// ═════════════════════════════════════════════════════════════════════════════

EngineContext::~EngineContext()
{
    Shutdown();
}

void error_callback(int error, const char* description)
{
    std::cerr << "Erro do GLFW [" << error << "]: " << description << std::endl;
}

bool EngineContext::Initialize(const char* windowTitle, uint32_t width, uint32_t height)
{
    glfwSetErrorCallback(error_callback);
    m_windowHeight = height;
    m_windowWidth = width;

    VK_LOG_INFO("Janela criada: " << width << "x" << height);

    if (!glfwInit()) {
        VK_LOG_ERROR("Falha ao inicializar GLFW");
        return false;
    }

    glfwWindowHint(GLFW_CLIENT_API, GLFW_NO_API);
    glfwWindowHint(GLFW_RESIZABLE, GLFW_TRUE);

    m_window = glfwCreateWindow(
        static_cast<int>(width), static_cast<int>(height), windowTitle, nullptr, nullptr);
    glfwSetWindowUserPointer(m_window, this);

    if (!m_window) {
        VK_LOG_ERROR("Falha ao criar janela GLFW");
        glfwTerminate();
        return false;
    }

    if (!InitializeVulkan()) {
        VK_LOG_ERROR("Falha ao inicializar Vulkan");
        glfwDestroyWindow(m_window);
        glfwTerminate();
        return false;
    }

    if (!InitializeSwapchain()) {
        VK_LOG_ERROR("Falha ao inicializar Swapchain");
        Shutdown();
        return false;
    }

    if (!InitializeRenderPass()) {
        VK_LOG_ERROR("Falha ao inicializar Renderpass");
        Shutdown();
        return false;
    }

    if (!InitializeFrameSyncing()) {
        VK_LOG_ERROR("Falha ao inicializar Frame Syncing");
        Shutdown();
        return false;
    }

    VK_LOG_INFO("EngineContext inicializado com sucesso!");
    return true;
}

bool EngineContext::InitializeVulkan()
{
    using namespace vkb;

    // ─────────────────────────────────────────────────────────────────────────
    // 1. Criar VkInstance
    // ─────────────────────────────────────────────────────────────────────────

    VK_LOG_INFO("Criando VkInstance...");

    InstanceBuilder instanceBuilder;
    auto inst_ret = instanceBuilder.set_app_name("Example")
                        .require_api_version(1, 3, 0)
                        .request_validation_layers()
                        .use_default_debug_messenger()
                        .build();

    if (!inst_ret) {
        VK_LOG_ERROR("Falha ao criar VkInstance: " << inst_ret.error().message());
        return false;
    }

    vkb::Instance vkb_inst = inst_ret.value();
    m_instance = vkb_inst.instance;
    m_debugMessenger = vkb_inst.debug_messenger;

    VK_LOG_INFO("✓ VkInstance criada");

    VK_LOG_INFO("  m_instance ptr : " << (void*)m_instance);

    // ─────────────────────────────────────────────────────────────────────────
    // 2. Criar Surface
    // ─────────────────────────────────────────────────────────────────────────

    VK_LOG_INFO("Criando VkSurfaceKHR...");

    VK_LOG_INFO("  m_window ptr   : " << (void*)m_window);
    VK_LOG_INFO("  glfwInit ok?   : " << glfwVulkanSupported());

    if (m_window == nullptr) {
        VK_LOG_ERROR("m_window é NULL antes de glfwCreateWindowSurface!");
        return false;
    }

    const char* glfwError = nullptr;
    int glfwErrCode = glfwGetError(&glfwError);
    if (glfwErrCode != GLFW_NO_ERROR) {
        VK_LOG_ERROR("Erro GLFW pendente antes da surface: ["
                     << glfwErrCode << "] " << (glfwError ? glfwError : "unknown"));
    }

    VkResult surfaceResult = glfwCreateWindowSurface(m_instance, m_window, nullptr, &m_surface);
    if (surfaceResult != VK_SUCCESS) {
        VK_LOG_ERROR("Falha ao criar VkSurfaceKHR: " << static_cast<int>(surfaceResult));
        return false;
    }

    VK_LOG_INFO("✓ VkSurfaceKHR criada");

    // ─────────────────────────────────────────────────────────────────────────
    // 3. Selecionar Physical Device
    // ─────────────────────────────────────────────────────────────────────────

    VK_LOG_INFO("Selecionando Physical Device...");

    PhysicalDeviceSelector selector{vkb_inst};
    auto phys_ret = selector.set_surface(m_surface).set_minimum_version(1, 3).select();

    if (!phys_ret) {
        VK_LOG_ERROR("Falha ao selecionar Physical Device: " << phys_ret.error().message());
        return false;
    }

    m_physicalDevice = phys_ret.value().physical_device;

    VK_LOG_INFO("✓ Physical Device: " << phys_ret.value().name);

    // ─────────────────────────────────────────────────────────────────────────
    // 4. Encontrar graphics queue family (manualmente)
    // ─────────────────────────────────────────────────────────────────────────

    uint32_t queueFamilyCount = 0;
    vkGetPhysicalDeviceQueueFamilyProperties(m_physicalDevice, &queueFamilyCount, nullptr);

    std::vector<VkQueueFamilyProperties> queueFamilies(queueFamilyCount);
    vkGetPhysicalDeviceQueueFamilyProperties(
        m_physicalDevice, &queueFamilyCount, queueFamilies.data());

    m_graphicsQueueFamily = 0;
    for (uint32_t i = 0; i < queueFamilies.size(); ++i) {
        if (queueFamilies[i].queueFlags & VK_QUEUE_GRAPHICS_BIT) {
            m_graphicsQueueFamily = i;
            break;
        }
    }

    VK_LOG_INFO("Graphics queue family: " << m_graphicsQueueFamily);

    // ─────────────────────────────────────────────────────────────────────────
    // 5. Criar Logical Device
    // ─────────────────────────────────────────────────────────────────────────

    VK_LOG_INFO("Criando VkDevice...");

    DeviceBuilder device_builder{phys_ret.value()};
    auto dev_ret = device_builder.build();

    if (!dev_ret) {
        VK_LOG_ERROR("Falha ao criar VkDevice: " << dev_ret.error().message());
        return false;
    }

    vkb::Device vkb_device = dev_ret.value();
    m_device = vkb_device.device;

    VK_LOG_INFO("✓ VkDevice criada");

    // ─────────────────────────────────────────────────────────────────────────
    // 6. Obter Graphics Queue
    // ─────────────────────────────────────────────────────────────────────────

    auto graphics_queue_ret = vkb_device.get_queue(vkb::QueueType::graphics);
    if (!graphics_queue_ret) {
        VK_LOG_ERROR("Falha ao obter graphics queue");
        return false;
    }

    m_graphicsQueue = graphics_queue_ret.value();

    VK_LOG_INFO("✓ Graphics Queue obtida");

    return true;
}

bool EngineContext::InitializeSwapchain()
{
    using namespace vkb;

    SwapchainBuilder swapchainBuilder{m_physicalDevice, m_device, m_surface};
    swapchainBuilder
        .set_desired_format({VK_FORMAT_B8G8R8A8_SRGB, VK_COLOR_SPACE_SRGB_NONLINEAR_KHR})
        .set_desired_present_mode(VK_PRESENT_MODE_FIFO_KHR)  // V-Sync
        .set_image_usage_flags(VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT);

    auto swapchainResult = swapchainBuilder.build();
    if (!swapchainResult) {
        VK_LOG_ERROR("Falha ao criar Swapchain: " << swapchainResult.error().message());
        return false;
    }

    auto swapchain = swapchainResult.value();
    m_swapchain = swapchain.swapchain;
    m_swapchainImages = swapchain.get_images().value();
    m_swapchainImageViews = swapchain.get_image_views().value();
    m_swapchainExtent = swapchain.extent;
    m_swapchainFormat = swapchain.image_format;

    VK_LOG_INFO("Swapchain criada: " << m_swapchainExtent.width << "x" << m_swapchainExtent.height
                                     << " (" << m_swapchainImages.size() << " imagens)");

    m_framebuffers.resize(m_swapchainImageViews.size());

    return true;
}

bool EngineContext::InitializeRenderPass()
{
    VkAttachmentDescription colorAttachment{};
    colorAttachment.format = m_swapchainFormat;
    colorAttachment.samples = VK_SAMPLE_COUNT_1_BIT;
    colorAttachment.loadOp = VK_ATTACHMENT_LOAD_OP_CLEAR;
    colorAttachment.storeOp = VK_ATTACHMENT_STORE_OP_STORE;
    colorAttachment.stencilLoadOp = VK_ATTACHMENT_LOAD_OP_DONT_CARE;
    colorAttachment.stencilStoreOp = VK_ATTACHMENT_STORE_OP_DONT_CARE;
    colorAttachment.initialLayout = VK_IMAGE_LAYOUT_UNDEFINED;
    colorAttachment.finalLayout = VK_IMAGE_LAYOUT_PRESENT_SRC_KHR;

    VkAttachmentReference colorAttachmentRef{};
    colorAttachmentRef.attachment = 0;
    colorAttachmentRef.layout = VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL;

    VkSubpassDescription subpass{};
    subpass.pipelineBindPoint = VK_PIPELINE_BIND_POINT_GRAPHICS;
    subpass.colorAttachmentCount = 1;
    subpass.pColorAttachments = &colorAttachmentRef;

    VkSubpassDependency dependency{};
    dependency.srcSubpass = VK_SUBPASS_EXTERNAL;
    dependency.dstSubpass = 0;
    dependency.srcStageMask = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
    dependency.dstStageMask = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
    dependency.srcAccessMask = 0;
    dependency.dstAccessMask = VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT;
    dependency.dependencyFlags = 0;

    VkRenderPassCreateInfo renderpassInfo{};
    renderpassInfo.sType = VK_STRUCTURE_TYPE_RENDER_PASS_CREATE_INFO;
    renderpassInfo.attachmentCount = 1;
    renderpassInfo.pAttachments = &colorAttachment;
    renderpassInfo.subpassCount = 1;
    renderpassInfo.pSubpasses = &subpass;
    renderpassInfo.dependencyCount = 1;
    renderpassInfo.pDependencies = &dependency;

    VK_CHECK(vkCreateRenderPass(m_device, &renderpassInfo, nullptr, &m_renderPass));

    VK_LOG_INFO("Renderpass criada com sucesso");

    for (size_t i = 0; i < m_swapchainImageViews.size(); ++i) {
        VkFramebufferCreateInfo framebufferInfo{};
        framebufferInfo.sType = VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO;
        framebufferInfo.renderPass = m_renderPass;
        framebufferInfo.attachmentCount = 1;
        framebufferInfo.pAttachments = &m_swapchainImageViews[i];
        framebufferInfo.width = m_swapchainExtent.width;
        framebufferInfo.height = m_swapchainExtent.height;
        framebufferInfo.layers = 1;

        VK_CHECK(vkCreateFramebuffer(m_device, &framebufferInfo, nullptr, &m_framebuffers[i]));
    }

    VK_LOG_INFO("Framebuffers criados: " << m_framebuffers.size());

    return true;
}

bool EngineContext::InitializeFrameSyncing()
{
    VkCommandPoolCreateInfo poolInfo{};
    poolInfo.sType = VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO;
    poolInfo.flags = VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT;
    poolInfo.queueFamilyIndex = m_graphicsQueueFamily;
    VK_CHECK(vkCreateCommandPool(m_device, &poolInfo, nullptr, &m_commandPool));
    VK_LOG_INFO("Command Pool criada");

    // Um FrameData por slot "em voo" (fence + acquire semaphore + command buffer)
    m_frames.resize(FRAMES_IN_FLIGHT);
    for (uint32_t i = 0; i < FRAMES_IN_FLIGHT; ++i)
        m_frames[i].Initialize(m_device, m_commandPool);

    // Um renderSemaphore por IMAGEM da swapchain — corrige VUID-vkQueueSubmit-00067
    VkSemaphoreCreateInfo semInfo{};
    semInfo.sType = VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO;
    m_renderSemaphores.resize(m_swapchainImages.size());
    for (auto& sem : m_renderSemaphores)
        VK_CHECK(vkCreateSemaphore(m_device, &semInfo, nullptr, &sem));

    VK_LOG_INFO("Frame Syncing inicializado — " << FRAMES_IN_FLIGHT << " frames in flight, "
                                                << m_renderSemaphores.size()
                                                << " render semaphores (um por imagem)");
    return true;
}

bool EngineContext::IsRunning() const
{
    return !glfwWindowShouldClose(m_window);
}

FrameData& EngineContext::BeginFrame()
{
    m_frameReady = false;
    m_currentFrame = (m_currentFrame + 1) % FRAMES_IN_FLIGHT;

    FrameData& frame = m_frames[m_currentFrame];
    VkFence renderFence = frame.GetRenderFence();
    VkSemaphore acquireSem = frame.GetAcquireSemaphore();
    VkCommandBuffer cmd = frame.GetCommandBuffer();

    VK_CHECK(vkWaitForFences(m_device, 1, &renderFence, VK_TRUE, UINT64_MAX));

    VkResult acqResult = vkAcquireNextImageKHR(
        m_device, m_swapchain, UINT64_MAX, acquireSem, VK_NULL_HANDLE, &m_swapchainImageIndex);

    if (acqResult == VK_ERROR_OUT_OF_DATE_KHR) {
        VK_LOG_WARN("Swapchain out of date");
        RecreateSwapchain();
        // m_frameReady permanece false → EndFrame() não fará nada
        return frame;
    }
    if (acqResult != VK_SUCCESS && acqResult != VK_SUBOPTIMAL_KHR) {
        VK_LOG_ERROR("Falha ao adquirir imagem da swapchain: " << static_cast<int>(acqResult));
        return frame;
    }

    VK_CHECK(vkResetFences(m_device, 1, &renderFence));
    VK_CHECK(vkResetCommandBuffer(cmd, 0));

    VkCommandBufferBeginInfo beginInfo{};
    beginInfo.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO;
    VK_CHECK(vkBeginCommandBuffer(cmd, &beginInfo));

    VkClearValue clearValue{{{0.1f, 0.1f, 0.1f, 1.0f}}};
    VkRenderPassBeginInfo renderPassInfo{};
    renderPassInfo.sType = VK_STRUCTURE_TYPE_RENDER_PASS_BEGIN_INFO;
    renderPassInfo.renderPass = m_renderPass;
    renderPassInfo.framebuffer = m_framebuffers[m_swapchainImageIndex];
    renderPassInfo.renderArea.extent = m_swapchainExtent;
    renderPassInfo.clearValueCount = 1;
    renderPassInfo.pClearValues = &clearValue;

    vkCmdBeginRenderPass(cmd, &renderPassInfo, VK_SUBPASS_CONTENTS_INLINE);

    m_frameReady = true;
    return frame;
}

void EngineContext::EndFrame()
{
    glfwPollEvents();

    if (!m_frameReady) {
        // BeginFrame() falhou (swapchain out of date) — não há nada para submeter
        return;
    }

    FrameData& frame = m_frames[m_currentFrame];
    VkCommandBuffer cmd = frame.GetCommandBuffer();
    VkSemaphore acquireSem = frame.GetAcquireSemaphore();
    VkSemaphore renderSem = m_renderSemaphores[m_swapchainImageIndex];  // por imagem!
    VkFence renderFence = frame.GetRenderFence();

    vkCmdEndRenderPass(cmd);
    VK_CHECK(vkEndCommandBuffer(cmd));

    VkPipelineStageFlags waitStage = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
    VkSubmitInfo submitInfo{};
    submitInfo.sType = VK_STRUCTURE_TYPE_SUBMIT_INFO;
    submitInfo.waitSemaphoreCount = 1;
    submitInfo.pWaitSemaphores = &acquireSem;  // espera a imagem estar disponível
    submitInfo.pWaitDstStageMask = &waitStage;
    submitInfo.commandBufferCount = 1;
    submitInfo.pCommandBuffers = &cmd;
    submitInfo.signalSemaphoreCount = 1;
    submitInfo.pSignalSemaphores = &renderSem;  // sinaliza quando render termina
    VK_CHECK(vkQueueSubmit(m_graphicsQueue, 1, &submitInfo, renderFence));

    VkPresentInfoKHR presentInfo{};
    presentInfo.sType = VK_STRUCTURE_TYPE_PRESENT_INFO_KHR;
    presentInfo.waitSemaphoreCount = 1;
    presentInfo.pWaitSemaphores = &renderSem;  // espera o render antes de apresentar
    presentInfo.swapchainCount = 1;
    presentInfo.pSwapchains = &m_swapchain;
    presentInfo.pImageIndices = &m_swapchainImageIndex;

    VkResult presentRes = vkQueuePresentKHR(m_graphicsQueue, &presentInfo);
    if (presentRes == VK_ERROR_OUT_OF_DATE_KHR || presentRes == VK_SUBOPTIMAL_KHR) {
        VK_LOG_WARN("Swapchain presentation outdated — recreate needed");
        RecreateSwapchain();
    }
}

void EngineContext::DestroyFrameSyncing()
{
    for (auto& frame : m_frames)
        frame.Destroy(m_device, m_commandPool);
    m_frames.clear();

    for (auto& sem : m_renderSemaphores) {
        if (sem != VK_NULL_HANDLE)
            vkDestroySemaphore(m_device, sem, nullptr);
    }
    m_renderSemaphores.clear();

    if (m_commandPool != VK_NULL_HANDLE) {
        vkDestroyCommandPool(m_device, m_commandPool, nullptr);
        m_commandPool = VK_NULL_HANDLE;
    }
}

void EngineContext::DestroySwapchain()
{
    if (m_device == VK_NULL_HANDLE)
        return;

    for (auto& fb : m_framebuffers) {
        if (fb != VK_NULL_HANDLE)
            vkDestroyFramebuffer(m_device, fb, nullptr);
    }
    m_framebuffers.clear();

    for (auto& view : m_swapchainImageViews) {
        if (view != VK_NULL_HANDLE)
            vkDestroyImageView(m_device, view, nullptr);
    }
    m_swapchainImageViews.clear();
    m_swapchainImages.clear();

    if (m_swapchain != VK_NULL_HANDLE) {
        vkDestroySwapchainKHR(m_device, m_swapchain, nullptr);
        m_swapchain = VK_NULL_HANDLE;
    }
}

void EngineContext::RecreateSwapchain()
{
    // Esperar enquanto a janela está minimizada (dimensões = 0)
    int width = 0, height = 0;
    glfwGetFramebufferSize(m_window, &width, &height);
    while (width == 0 || height == 0) {
        glfwGetFramebufferSize(m_window, &width, &height);
        glfwWaitEvents();
    }

    // Garantir que a GPU terminou todo trabalho pendente
    vkDeviceWaitIdle(m_device);

    // ── Destruir recursos dependentes da swapchain ────────────────────────
    for (auto& fb : m_framebuffers) {
        if (fb != VK_NULL_HANDLE)
            vkDestroyFramebuffer(m_device, fb, nullptr);
    }
    m_framebuffers.clear();

    for (auto& view : m_swapchainImageViews) {
        if (view != VK_NULL_HANDLE)
            vkDestroyImageView(m_device, view, nullptr);
    }
    m_swapchainImageViews.clear();
    m_swapchainImages.clear();

    // Semáforos de render dependem da contagem de imagens — podem mudar
    for (auto& sem : m_renderSemaphores) {
        if (sem != VK_NULL_HANDLE)
            vkDestroySemaphore(m_device, sem, nullptr);
    }
    m_renderSemaphores.clear();

    // ── Recriar Swapchain ─────────────────────────────────────────────────
    // Guardamos o handle antigo para passar ao builder (permite reuso interno pelo driver)
    VkSwapchainKHR oldSwapchain = m_swapchain;
    m_swapchain = VK_NULL_HANDLE;

    using namespace vkb;
    SwapchainBuilder swapchainBuilder{m_physicalDevice, m_device, m_surface};
    auto swapchainResult =
        swapchainBuilder
            .set_desired_format({VK_FORMAT_B8G8R8A8_SRGB, VK_COLOR_SPACE_SRGB_NONLINEAR_KHR})
            .set_desired_present_mode(VK_PRESENT_MODE_FIFO_KHR)
            .set_image_usage_flags(VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT)
            .set_old_swapchain(oldSwapchain)  // ← chave para recriação eficiente
            .build();

    // Agora sim podemos destruir a swapchain antiga
    if (oldSwapchain != VK_NULL_HANDLE)
        vkDestroySwapchainKHR(m_device, oldSwapchain, nullptr);

    if (!swapchainResult) {
        VK_LOG_ERROR("Falha ao recriar Swapchain: " << swapchainResult.error().message());
        return;
    }

    auto swapchain = swapchainResult.value();
    m_swapchain = swapchain.swapchain;
    m_swapchainImages = swapchain.get_images().value();
    m_swapchainImageViews = swapchain.get_image_views().value();
    m_swapchainExtent = swapchain.extent;
    m_swapchainFormat = swapchain.image_format;

    VK_LOG_INFO("Swapchain recriada: " << m_swapchainExtent.width << "x" << m_swapchainExtent.height
                                       << " (" << m_swapchainImages.size() << " imagens)");

    // ── Recriar Framebuffers ──────────────────────────────────────────────
    m_framebuffers.resize(m_swapchainImageViews.size());
    for (size_t i = 0; i < m_swapchainImageViews.size(); ++i) {
        VkFramebufferCreateInfo fbInfo{};
        fbInfo.sType = VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO;
        fbInfo.renderPass = m_renderPass;  // render pass pode ser reutilizado
        fbInfo.attachmentCount = 1;
        fbInfo.pAttachments = &m_swapchainImageViews[i];
        fbInfo.width = m_swapchainExtent.width;
        fbInfo.height = m_swapchainExtent.height;
        fbInfo.layers = 1;
        VK_CHECK(vkCreateFramebuffer(m_device, &fbInfo, nullptr, &m_framebuffers[i]));
    }

    // ── Recriar semáforos de render (um por imagem) ───────────────────────
    VkSemaphoreCreateInfo semInfo{};
    semInfo.sType = VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO;
    m_renderSemaphores.resize(m_swapchainImages.size());
    for (auto& sem : m_renderSemaphores)
        VK_CHECK(vkCreateSemaphore(m_device, &semInfo, nullptr, &sem));

    VK_LOG_INFO("Framebuffers e semáforos recriados com sucesso");
}

void EngineContext::Shutdown()
{
    if (m_device == VK_NULL_HANDLE)
        return;

    vkDeviceWaitIdle(m_device);

    DestroyFrameSyncing();

    if (m_renderPass != VK_NULL_HANDLE) {
        vkDestroyRenderPass(m_device, m_renderPass, nullptr);
        m_renderPass = VK_NULL_HANDLE;
    }

    DestroySwapchain();

    if (m_device != VK_NULL_HANDLE) {
        vkDestroyDevice(m_device, nullptr);
        m_device = VK_NULL_HANDLE;
    }

    if (m_surface != VK_NULL_HANDLE) {
        vkDestroySurfaceKHR(m_instance, m_surface, nullptr);
        m_surface = VK_NULL_HANDLE;
    }

    if (m_debugMessenger != VK_NULL_HANDLE) {
        auto destroyFn = reinterpret_cast<PFN_vkDestroyDebugUtilsMessengerEXT>(
            vkGetInstanceProcAddr(m_instance, "vkDestroyDebugUtilsMessengerEXT"));
        if (destroyFn)
            destroyFn(m_instance, m_debugMessenger, nullptr);
        m_debugMessenger = VK_NULL_HANDLE;
    }

    if (m_instance != VK_NULL_HANDLE) {
        vkDestroyInstance(m_instance, nullptr);
        m_instance = VK_NULL_HANDLE;
    }

    if (m_window != nullptr) {
        glfwDestroyWindow(m_window);
        m_window = nullptr;
    }

    glfwTerminate();
    VK_LOG_INFO("EngineContext desligado com sucesso");
}

}  // namespace vklib
