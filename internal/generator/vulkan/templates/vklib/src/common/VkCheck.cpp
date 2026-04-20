#include <vklib/common/VkCheck.h>
#include <sstream>
#include <iostream>

namespace vklib {
    std::string_view VkResultToString(VkResult result) {
        switch (result) {
            case VK_SUCCESS:                        return "VK_SUCCESS";
            case VK_NOT_READY:                      return "VK_NOT_READY";
            case VK_TIMEOUT:                        return "VK_TIMEOUT";
            case VK_EVENT_SET:                      return "VK_EVENT_SET";
            case VK_EVENT_RESET:                    return "VK_EVENT_RESET";
            case VK_INCOMPLETE:                     return "VK_INCOMPLETE";
            case VK_ERROR_OUT_OF_HOST_MEMORY:       return "ERROR_OUT_OF_HOST_MEMORY";
            case VK_ERROR_OUT_OF_DEVICE_MEMORY:     return "ERROR_OUT_OF_DEVICE_MEMORY";
            case VK_ERROR_INITIALIZATION_FAILED:    return "ERROR_INITIALIZATION_FAILED";
            case VK_ERROR_DEVICE_LOST:              return "ERROR_DEVICE_LOST";
            case VK_ERROR_MEMORY_MAP_FAILED:        return "ERROR_MEMORY_MAP_FAILED";
            case VK_ERROR_LAYER_NOT_PRESENT:        return "ERROR_LAYER_NOT_PRESENT";
            case VK_ERROR_EXTENSION_NOT_PRESENT:    return "ERROR_EXTENSION_NOT_PRESENT";
            case VK_ERROR_FEATURE_NOT_PRESENT:      return "ERROR_FEATURE_NOT_PRESENT";
            case VK_ERROR_INCOMPATIBLE_DRIVER:      return "ERROR_INCOMPATIBLE_DRIVER";
            case VK_ERROR_TOO_MANY_OBJECTS:         return "ERROR_TOO_MANY_OBJECTS";
            case VK_ERROR_FORMAT_NOT_SUPPORTED:     return "ERROR_FORMAT_NOT_SUPPORTED";
            case VK_ERROR_FRAGMENTED_POOL:          return "ERROR_FRAGMENTED_POOL";
            case VK_ERROR_UNKNOWN:                  return "ERROR_UNKNOWN";
            case VK_ERROR_OUT_OF_POOL_MEMORY:       return "ERROR_OUT_OF_POOL_MEMORY";
            case VK_ERROR_INVALID_EXTERNAL_HANDLE:  return "ERROR_INVALID_EXTERNAL_HANDLE";
            case VK_ERROR_FRAGMENTATION:            return "ERROR_FRAGMENTATION";
            case VK_ERROR_INVALID_OPAQUE_CAPTURE_ADDRESS: return "ERROR_INVALID_OPAQUE_CAPTURE_ADDRESS";
            case VK_ERROR_SURFACE_LOST_KHR:         return "ERROR_SURFACE_LOST_KHR";
            case VK_ERROR_NATIVE_WINDOW_IN_USE_KHR: return "ERROR_NATIVE_WINDOW_IN_USE_KHR";
            case VK_SUBOPTIMAL_KHR:                 return "SUBOPTIMAL_KHR";
            case VK_ERROR_OUT_OF_DATE_KHR:          return "ERROR_OUT_OF_DATE_KHR";
            default:                                 return "UNKNOWN_ERROR";
        }
    }

    void LogVulkanError(VkResult result, const char* file, int line,
                        std::string_view context) {
                            std::cerr << "== Vulkan Error in " << context << "\n"
                                      << "   File: " << file << ":" << line << "\n"
                                      << "   Code: " << static_cast<int>(result) << " ("
                                      << VkResultToString(result) << ") ==\n";
    }
}
