#pragma once

/**
 * @file VulkanLib.h
 * @brief Header principal que inclui toda a biblioteca VulkanLib
 */

// ═════════════════════════════════════════════════════════════════════════════
// Common
// ═════════════════════════════════════════════════════════════════════════════

#include "common/Defines.h"
#include "common/Types.h"
#include "common/VkCheck.h"

// ═════════════════════════════════════════════════════════════════════════════
// Core
// ═════════════════════════════════════════════════════════════════════════════

#include "core/EngineContext.h"
#include "core/FrameData.h"

// ═════════════════════════════════════════════════════════════════════════════
// Memory
// ═════════════════════════════════════════════════════════════════════════════

#include "memory/Buffer.h"
#include "memory/Image.h"

// ═════════════════════════════════════════════════════════════════════════════
// Pipeline
// ═════════════════════════════════════════════════════════════════════════════

#include "pipeline/Pipeline.h"
#include "pipeline/PipelineBuilder.h"
#include "pipeline/ShaderLoader.h"

// ═════════════════════════════════════════════════════════════════════════════
// Descriptors
// ═════════════════════════════════════════════════════════════════════════════

#include "descriptors/DescriptorAllocator.h"
#include "descriptors/DescriptorWriter.h"
