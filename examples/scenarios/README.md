# Real-World Scenario Examples

Production-ready testing scenarios using the enhanced VasDeference framework for complete business use cases including e-commerce, financial services, IoT, and ML pipelines.

## Overview

These examples demonstrate real-world, production-tested patterns for complete business scenarios:

- **E-Commerce Platform**: Complete order processing, inventory, and fulfillment pipeline
- **Financial Services**: Transaction processing, fraud detection, and compliance
- **IoT Data Processing**: Device data ingestion, processing, and analytics
- **ML Model Pipeline**: Training, deployment, and inference workflows
- **Content Management**: Media processing, storage, and delivery
- **Healthcare**: Patient data processing with HIPAA compliance

## Examples

### 1. E-Commerce Platform
**File**: `ecommerce_platform_test.go`

Complete e-commerce testing scenario:
- Product catalog management
- Shopping cart and checkout
- Order processing pipeline
- Inventory management
- Payment processing
- Shipping and fulfillment
- Customer notifications
- Analytics and reporting

### 2. Financial Transaction Processing
**File**: `financial_transaction_test.go`

Financial services testing:
- Real-time transaction processing
- Fraud detection and prevention
- Risk assessment workflows
- Compliance reporting
- Account management
- Payment settlement
- Audit trail generation

### 3. IoT Data Processing Pipeline
**File**: `iot_data_pipeline_test.go`

IoT platform testing:
- Device data ingestion
- Stream processing
- Real-time analytics
- Anomaly detection
- Data archival
- Device management
- Alert generation

### 4. ML Model Deployment Pipeline
**File**: `ml_model_pipeline_test.go`

Machine learning workflow testing:
- Model training pipeline
- Model validation and testing
- A/B testing framework
- Model deployment
- Inference API
- Performance monitoring
- Model rollback

### 5. Healthcare Data Processing
**File**: `healthcare_data_test.go`

Healthcare system testing:
- Patient data ingestion
- HIPAA compliance validation
- Clinical workflow processing
- Medical record management
- Appointment scheduling
- Insurance processing
- Audit trail compliance

## Key Features

### Complete Business Workflows
```go
// E-commerce order processing with full pipeline testing
func TestCompleteOrderProcessingWorkflow(t *testing.T) {
    ctx := setupECommerceInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Create test order with real business logic
    order := createTestOrder(t, OrderRequest{
        CustomerId: "cust-12345",
        Items: []OrderItem{
            {ProductId: "prod-001", Quantity: 2, Price: 29.99},
            {ProductId: "prod-002", Quantity: 1, Price: 49.99},
        },
        ShippingAddress: validShippingAddress,
        PaymentMethod: validPaymentMethod,
    })
    
    // Execute complete workflow
    result := processCompleteOrder(t, ctx, order)
    
    // Validate all aspects of the business process
    validateOrderCreation(t, result.OrderCreated)
    validateInventoryReservation(t, result.InventoryReserved)
    validatePaymentProcessing(t, result.PaymentProcessed)
    validateShippingArrangement(t, result.ShippingArranged)
    validateCustomerNotifications(t, result.NotificationsSent)
    validateAnalyticsEvents(t, result.AnalyticsRecorded)
}
```

### Real-World Data Patterns
```go
// Financial transaction with realistic data volumes
func TestHighVolumeTransactionProcessing(t *testing.T) {
    ctx := setupFinancialInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Generate realistic transaction patterns
    transactions := generateRealisticTransactionPatterns(TransactionConfig{
        TotalTransactions:    100000,
        TransactionsPerSecond: 1000,
        FraudRate:            0.02, // 2% fraud
        AverageAmount:        75.50,
        CurrencyDistribution: map[string]float64{
            "USD": 0.70,
            "EUR": 0.20,
            "GBP": 0.10,
        },
    })
    
    // Process transactions with parallel table testing
    vasdeference.TableTest[Transaction, TransactionResult](t, "High Volume Processing").
        Parallel().
        WithMaxWorkers(50).
        Run(func(tx Transaction) TransactionResult {
            return processTransaction(t, ctx, tx)
        }, validateTransactionProcessing)
}
```

### Industry-Specific Compliance
```go
// HIPAA compliant healthcare data processing
func TestHIPAACompliantDataProcessing(t *testing.T) {
    ctx := setupHealthcareInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Test with PHI (Protected Health Information)
    patientData := createTestPHI(t, PatientRecord{
        PatientId:       "patient-001",
        Name:           "Test Patient",
        DateOfBirth:    "1985-01-15",
        MedicalHistory: "Type 2 Diabetes, Hypertension",
        InsuranceInfo:  validInsuranceInfo,
    })
    
    // Process with HIPAA compliance validation
    result := processPatientData(t, ctx, patientData)
    
    // Validate HIPAA compliance at every step
    validateDataEncryption(t, result.EncryptedData)
    validateAccessControls(t, result.AccessLog)
    validateAuditTrail(t, result.AuditEntries)
    validateDataMinimization(t, result.ProcessedData)
    validateBreachNotification(t, result.SecurityEvents)
}
```

### Production-Scale Testing
```go
// IoT data processing with real-world scale
func TestIoTDataProcessingScale(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping large-scale IoT test")
    }
    
    ctx := setupIoTInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Simulate production IoT data volumes
    deviceCount := 10000
    messagesPerDevicePerMinute := 60
    testDuration := 15 * time.Minute
    
    // Generate realistic IoT data patterns
    dataStream := generateIoTDataStream(t, IoTStreamConfig{
        DeviceCount:    deviceCount,
        MessageRate:    messagesPerDevicePerMinute,
        Duration:       testDuration,
        DeviceTypes:    []string{"sensor", "gateway", "actuator"},
        DataTypes:      []string{"temperature", "humidity", "pressure", "motion"},
        AnomalyRate:    0.001, // 0.1% anomalies
    })
    
    // Process stream with performance validation
    result := processIoTDataStream(t, ctx, dataStream)
    
    expectedMessages := deviceCount * messagesPerDevicePerMinute * int(testDuration.Minutes())
    assert.GreaterOrEqual(t, result.ProcessedMessages, expectedMessages*95/100, 
        "Should process at least 95% of messages")
    assert.LessOrEqual(t, result.ProcessingLatency, 5*time.Second,
        "Processing latency should be under 5 seconds")
}
```

## Running Scenario Tests

### Prerequisites
```bash
# Set scenario-specific configuration
export SCENARIO_TEST_MODE=true
export ENABLE_REAL_WORLD_DATA=true
export DATA_VOLUME_MULTIPLIER=1

# Configure business-specific settings
export ECOMMERCE_CATALOG_SIZE=10000
export FINANCIAL_TPS_LIMIT=5000
export IOT_DEVICE_COUNT=10000
export ML_MODEL_TRAINING_DATA=large

# Compliance settings
export ENABLE_HIPAA_COMPLIANCE=true
export ENABLE_PCI_DSS_COMPLIANCE=true
export ENABLE_GDPR_COMPLIANCE=true
```

### Execute Tests
```bash
# Run all scenario tests
go test -v -timeout=3600s ./examples/scenarios/

# Run specific business scenarios
go test -v ./examples/scenarios/ -run TestECommerce
go test -v ./examples/scenarios/ -run TestFinancial
go test -v ./examples/scenarios/ -run TestIoT
go test -v ./examples/scenarios/ -run TestMachineLearning

# Run with production data volumes
PRODUCTION_SCALE=true go test -v -timeout=7200s ./examples/scenarios/

# Run compliance-specific tests
COMPLIANCE_MODE=HIPAA go test -v ./examples/scenarios/ -run TestHealthcare
COMPLIANCE_MODE=PCI_DSS go test -v ./examples/scenarios/ -run TestFinancial
```

### Performance Testing
```bash
# Business scenario performance benchmarks
go test -v -bench=BenchmarkECommerce ./examples/scenarios/
go test -v -bench=BenchmarkFinancial ./examples/scenarios/
go test -v -bench=BenchmarkIoT ./examples/scenarios/

# Load testing with realistic patterns
LOAD_TEST=true go test -v -timeout=3600s ./examples/scenarios/
```

## Business Use Cases

### E-Commerce Platform Features

#### Product Catalog Management
```go
func TestProductCatalogManagement(t *testing.T) {
    // Product creation, updates, pricing, inventory tracking
    catalog := setupProductCatalog(t, CatalogConfig{
        ProductCount:    10000,
        Categories:      50,
        BrandsCount:     200,
        PriceRanges:     []PriceRange{{Min: 10, Max: 1000}},
        InventoryLevels: "realistic",
    })
    
    // Test catalog operations
    validateProductSearch(t, catalog)
    validatePriceCalculation(t, catalog)
    validateInventoryTracking(t, catalog)
    validateCategoryNavigation(t, catalog)
}
```

#### Order Fulfillment Pipeline
```go
func TestOrderFulfillmentPipeline(t *testing.T) {
    // Complete order lifecycle testing
    pipeline := setupFulfillmentPipeline(t)
    
    orderStates := []OrderState{
        "placed", "validated", "payment_processed", 
        "inventory_reserved", "picked", "packed", 
        "shipped", "delivered", "completed",
    }
    
    for _, state := range orderStates {
        validateOrderStateTransition(t, pipeline, state)
        validateBusinessRules(t, pipeline, state)
        validateNotifications(t, pipeline, state)
    }
}
```

### Financial Services Features

#### Fraud Detection System
```go
func TestFraudDetectionSystem(t *testing.T) {
    ctx := setupFinancialInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Test various fraud patterns
    fraudPatterns := []FraudPattern{
        {Type: "velocity", Description: "Too many transactions in short time"},
        {Type: "amount", Description: "Unusual transaction amount"},
        {Type: "location", Description: "Transaction from unusual location"},
        {Type: "merchant", Description: "High-risk merchant category"},
        {Type: "behavioral", Description: "Unusual spending pattern"},
    }
    
    for _, pattern := range fraudPatterns {
        t.Run(pattern.Type, func(t *testing.T) {
            transaction := generateFraudulentTransaction(pattern)
            result := processFraudDetection(t, ctx, transaction)
            
            assert.True(t, result.FraudDetected, "Should detect fraud pattern: %s", pattern.Type)
            assert.Greater(t, result.RiskScore, 0.8, "Risk score should be high for fraud")
            validateFraudResponse(t, result, pattern)
        })
    }
}
```

#### Regulatory Compliance
```go
func TestRegulatoryCompliance(t *testing.T) {
    regulations := []Regulation{
        {Name: "PCI-DSS", Requirements: []string{"card_data_encryption", "access_control"}},
        {Name: "SOX", Requirements: []string{"audit_trail", "segregation_of_duties"}},
        {Name: "GDPR", Requirements: []string{"data_protection", "right_to_be_forgotten"}},
    }
    
    for _, regulation := range regulations {
        validateRegulationCompliance(t, regulation)
    }
}
```

### IoT Platform Features

#### Real-Time Data Processing
```go
func TestIoTRealTimeProcessing(t *testing.T) {
    ctx := setupIoTInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Simulate real-time IoT data streams
    streams := []IoTStream{
        {DeviceType: "temperature_sensor", MessageRate: 60, DataFormat: "json"},
        {DeviceType: "pressure_sensor", MessageRate: 30, DataFormat: "binary"},
        {DeviceType: "motion_detector", MessageRate: 1, DataFormat: "event"},
    }
    
    for _, stream := range streams {
        t.Run(stream.DeviceType, func(t *testing.T) {
            processor := createStreamProcessor(t, ctx, stream)
            
            // Process data for 5 minutes
            result := processStreamForDuration(t, processor, 5*time.Minute)
            
            validateProcessingLatency(t, result)
            validateDataAccuracy(t, result)
            validateAnomalyDetection(t, result)
        })
    }
}
```

#### Device Management
```go
func TestIoTDeviceManagement(t *testing.T) {
    // Device lifecycle management
    deviceOperations := []DeviceOperation{
        "provision", "configure", "monitor", "update", "troubleshoot", "decommission",
    }
    
    for _, operation := range deviceOperations {
        validateDeviceOperation(t, operation)
    }
}
```

### ML Pipeline Features

#### Model Training Pipeline
```go
func TestMLModelTrainingPipeline(t *testing.T) {
    ctx := setupMLInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    trainingPipeline := MLTrainingPipeline{
        DataPreparation: "feature_engineering",
        ModelTypes:      []string{"linear_regression", "random_forest", "neural_network"},
        ValidationStrategy: "k_fold_cross_validation",
        HyperparameterTuning: true,
    }
    
    result := executeTrainingPipeline(t, ctx, trainingPipeline)
    
    validateModelAccuracy(t, result.Models)
    validateModelPerformance(t, result.Metrics)
    validateModelValidation(t, result.ValidationResults)
}
```

#### A/B Testing Framework
```go
func TestMLModelABTesting(t *testing.T) {
    // Test model A/B testing infrastructure
    abTest := MLABTest{
        ModelA: "current_production_model",
        ModelB: "new_candidate_model",
        TrafficSplit: 0.1, // 10% to model B
        SuccessMetric: "conversion_rate",
        MinimumSampleSize: 1000,
    }
    
    result := executeABTest(t, abTest, 7*24*time.Hour) // 7 days
    validateABTestResults(t, result)
}
```

## Monitoring and Observability

### Business Metrics
```go
// Real-world business metrics collection
type BusinessMetrics struct {
    // E-commerce
    ConversionRate    float64 `json:"conversionRate"`
    AverageOrderValue float64 `json:"averageOrderValue"`
    CustomerSatisfaction float64 `json:"customerSatisfaction"`
    
    // Financial
    TransactionSuccessRate float64 `json:"transactionSuccessRate"`
    FraudDetectionAccuracy float64 `json:"fraudDetectionAccuracy"`
    ComplianceScore       float64 `json:"complianceScore"`
    
    // IoT
    DeviceUptime         float64 `json:"deviceUptime"`
    DataProcessingLatency time.Duration `json:"dataProcessingLatency"`
    AnomalyDetectionRate float64 `json:"anomalyDetectionRate"`
    
    // ML
    ModelAccuracy     float64 `json:"modelAccuracy"`
    PredictionLatency time.Duration `json:"predictionLatency"`
    ModelDrift        float64 `json:"modelDrift"`
}

func collectBusinessMetrics(t *testing.T, scenario string) BusinessMetrics {
    // Implementation would collect actual business metrics
    return BusinessMetrics{}
}
```

### SLA Validation
```go
// Service Level Agreement validation
func TestBusinessSLACompliance(t *testing.T) {
    slas := []SLA{
        {Service: "order_processing", Metric: "latency", Threshold: 30 * time.Second, Target: 0.99},
        {Service: "payment_processing", Metric: "availability", Threshold: 99.9, Target: 1.0},
        {Service: "fraud_detection", Metric: "accuracy", Threshold: 0.95, Target: 1.0},
        {Service: "iot_data_processing", Metric: "throughput", Threshold: 10000, Target: 1.0},
    }
    
    for _, sla := range slas {
        validateSLACompliance(t, sla)
    }
}
```

### Cost Optimization
```go
// Business cost optimization testing
func TestCostOptimization(t *testing.T) {
    scenarios := []CostScenario{
        {Name: "peak_traffic", Multiplier: 5.0, Duration: 2 * time.Hour},
        {Name: "normal_traffic", Multiplier: 1.0, Duration: 22 * time.Hour},
        {Name: "low_traffic", Multiplier: 0.3, Duration: 4 * time.Hour},
    }
    
    for _, scenario := range scenarios {
        cost := calculateScenarioCost(t, scenario)
        validateCostEfficiency(t, cost, scenario)
    }
}
```

## Integration Testing

### Third-Party Services
```go
// Test integration with external services
func TestThirdPartyIntegrations(t *testing.T) {
    integrations := []ThirdPartyIntegration{
        {Service: "payment_gateway", Provider: "stripe", Critical: true},
        {Service: "shipping_provider", Provider: "fedex", Critical: true},
        {Service: "email_service", Provider: "sendgrid", Critical: false},
        {Service: "analytics", Provider: "google_analytics", Critical: false},
    }
    
    for _, integration := range integrations {
        validateIntegration(t, integration)
    }
}
```

### Legacy System Integration
```go
// Test integration with legacy systems
func TestLegacySystemIntegration(t *testing.T) {
    legacySystems := []LegacySystem{
        {Name: "mainframe_inventory", Protocol: "cobol", DataFormat: "fixed_width"},
        {Name: "legacy_crm", Protocol: "soap", DataFormat: "xml"},
        {Name: "accounting_system", Protocol: "odbc", DataFormat: "sql"},
    }
    
    for _, system := range legacySystems {
        validateLegacyIntegration(t, system)
    }
}
```

---

**Next**: Select a specific business scenario that matches your domain to implement comprehensive testing patterns.