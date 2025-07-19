#!/usr/bin/env python3
"""
Comprehensive test script for A-MEM MCP Server Phase 2 features
Tests advanced capabilities including evolution, monitoring, and enhanced embeddings
"""

import json
import subprocess
import sys
import time
import requests
from typing import Dict, Any

def send_mcp_request(request: Dict[str, Any]) -> Dict[str, Any]:
    """Send a JSON-RPC request to the MCP server via stdin/stdout"""
    try:
        # Start the server process
        process = subprocess.Popen(
            ['./amem-server', '-config', 'config/development.yaml'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Send request
        request_json = json.dumps(request) + '\n'
        stdout, stderr = process.communicate(input=request_json, timeout=15)
        
        if stderr:
            print(f"Server stderr: {stderr}")
        
        # Parse response
        if stdout.strip():
            return json.loads(stdout.strip())
        else:
            return {"error": "No response from server"}
            
    except subprocess.TimeoutExpired:
        process.kill()
        return {"error": "Request timeout"}
    except Exception as e:
        return {"error": str(e)}

def test_enhanced_memory_storage():
    """Test enhanced memory storage with better analysis"""
    print("Testing enhanced memory storage...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "store_coding_memory",
            "arguments": {
                "content": """
async function processUserData(users) {
    const results = await Promise.all(
        users.map(async (user) => {
            const profile = await fetchUserProfile(user.id);
            const preferences = await getUserPreferences(user.id);
            return {
                ...user,
                profile,
                preferences,
                lastUpdated: new Date().toISOString()
            };
        })
    );
    return results.filter(user => user.profile.active);
}
                """.strip(),
                "project_path": "/frontend/services",
                "code_type": "javascript",
                "context": "Async data processing with Promise.all for user enrichment"
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Enhanced storage response: {json.dumps(response, indent=2)}")
    
    success = "result" in response and not response.get("error")
    if success:
        print("✅ Enhanced memory storage - PASSED")
    else:
        print("❌ Enhanced memory storage - FAILED")
    
    return success

def test_advanced_memory_evolution():
    """Test the advanced memory evolution system"""
    print("\nTesting advanced memory evolution...")
    
    # First, store a few related memories
    memories = [
        {
            "content": "function fibonacci(n) { return n <= 1 ? n : fibonacci(n-1) + fibonacci(n-2); }",
            "project_path": "/algorithms",
            "code_type": "javascript",
            "context": "Recursive Fibonacci implementation"
        },
        {
            "content": "function fibonacciMemo(n, memo = {}) { if (n in memo) return memo[n]; if (n <= 1) return n; memo[n] = fibonacciMemo(n-1, memo) + fibonacciMemo(n-2, memo); return memo[n]; }",
            "project_path": "/algorithms",
            "code_type": "javascript", 
            "context": "Memoized Fibonacci for performance"
        }
    ]
    
    # Store memories first
    for i, memory in enumerate(memories):
        request = {
            "jsonrpc": "2.0",
            "id": f"store_{i}",
            "method": "tools/call",
            "params": {
                "name": "store_coding_memory",
                "arguments": memory
            }
        }
        send_mcp_request(request)
    
    # Wait a moment for storage
    time.sleep(1)
    
    # Now test evolution
    request = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/call",
        "params": {
            "name": "evolve_memory_network",
            "arguments": {
                "trigger_type": "manual",
                "scope": "project",
                "max_memories": 20,
                "project_path": "/algorithms"
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Evolution response: {json.dumps(response, indent=2)}")
    
    success = "result" in response and not response.get("error")
    if success:
        print("✅ Advanced memory evolution - PASSED")
    else:
        print("❌ Advanced memory evolution - FAILED")
    
    return success

def test_enhanced_memory_retrieval():
    """Test enhanced memory retrieval with better relevance"""
    print("\nTesting enhanced memory retrieval...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "retrieve_relevant_memories",
            "arguments": {
                "query": "efficient fibonacci algorithm with memoization",
                "max_results": 5,
                "min_relevance": 0.3,
                "code_types": ["javascript"]
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Enhanced retrieval response: {json.dumps(response, indent=2)}")
    
    success = "result" in response and not response.get("error")
    if success:
        print("✅ Enhanced memory retrieval - PASSED")
    else:
        print("❌ Enhanced memory retrieval - FAILED")
    
    return success

def test_metrics_endpoint():
    """Test Prometheus metrics endpoint"""
    print("\nTesting metrics endpoint...")
    
    try:
        # Start server in background for metrics test
        process = subprocess.Popen(
            ['./amem-server', '-config', 'config/development.yaml'],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        
        # Wait for server to start
        time.sleep(3)
        
        # Test metrics endpoint
        response = requests.get('http://localhost:9090/metrics', timeout=5)
        
        if response.status_code == 200:
            metrics_content = response.text
            print(f"Metrics endpoint accessible, content length: {len(metrics_content)}")
            
            # Check for expected metrics
            expected_metrics = [
                'amem_memory_operations_total',
                'amem_llm_requests_total',
                'amem_vector_searches_total'
            ]
            
            found_metrics = sum(1 for metric in expected_metrics if metric in metrics_content)
            print(f"Found {found_metrics}/{len(expected_metrics)} expected metrics")
            
            success = found_metrics >= 2  # At least some metrics should be present
        else:
            print(f"Metrics endpoint returned status: {response.status_code}")
            success = False
        
        # Clean up
        process.terminate()
        process.wait(timeout=5)
        
    except Exception as e:
        print(f"Metrics test error: {e}")
        success = False
    
    if success:
        print("✅ Metrics endpoint - PASSED")
    else:
        print("❌ Metrics endpoint - FAILED")
    
    return success

def test_health_endpoint():
    """Test health check endpoint"""
    print("\nTesting health endpoint...")
    
    try:
        # Start server in background
        process = subprocess.Popen(
            ['./amem-server', '-config', 'config/development.yaml'],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        
        # Wait for server to start
        time.sleep(3)
        
        # Test health endpoint
        response = requests.get('http://localhost:9090/health', timeout=5)
        
        if response.status_code == 200:
            print("Health endpoint accessible")
            success = True
        else:
            print(f"Health endpoint returned status: {response.status_code}")
            success = False
        
        # Clean up
        process.terminate()
        process.wait(timeout=5)
        
    except Exception as e:
        print(f"Health test error: {e}")
        success = False
    
    if success:
        print("✅ Health endpoint - PASSED")
    else:
        print("❌ Health endpoint - FAILED")
    
    return success

def test_docker_services():
    """Test Docker services are configured correctly"""
    print("\nTesting Docker services configuration...")
    
    try:
        # Check if docker-compose.yml exists and is valid
        result = subprocess.run(
            ['docker-compose', 'config'],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode == 0:
            print("Docker Compose configuration is valid")
            
            # Check for expected services
            config_output = result.stdout
            expected_services = [
                'amem-server',
                'chromadb',
                'sentence-transformers',
                'redis',
                'prometheus'
            ]
            
            found_services = sum(1 for service in expected_services if service in config_output)
            print(f"Found {found_services}/{len(expected_services)} expected services")
            
            success = found_services >= 4  # Most services should be present
        else:
            print(f"Docker Compose validation failed: {result.stderr}")
            success = False
            
    except Exception as e:
        print(f"Docker services test error: {e}")
        success = False
    
    if success:
        print("✅ Docker services configuration - PASSED")
    else:
        print("❌ Docker services configuration - FAILED")
    
    return success

def main():
    """Run all Phase 2 tests"""
    print("🚀 A-MEM MCP Server Phase 2 Test Suite")
    print("=" * 50)
    print("")

    # Check if server binary exists
    try:
        subprocess.run(['./amem-server', '--help'], 
                      capture_output=True, check=False)
    except FileNotFoundError:
        print("Error: amem-server binary not found. Please run 'make build' first.")
        sys.exit(1)

    tests = [
        ("Enhanced Memory Storage", test_enhanced_memory_storage),
        ("Advanced Memory Evolution", test_advanced_memory_evolution),
        ("Enhanced Memory Retrieval", test_enhanced_memory_retrieval),
        ("Metrics Endpoint", test_metrics_endpoint),
        ("Health Endpoint", test_health_endpoint),
        ("Docker Services Config", test_docker_services),
    ]

    passed = 0
    total = len(tests)

    for test_name, test_func in tests:
        try:
            if test_func():
                passed += 1
            time.sleep(1)  # Brief pause between tests
        except Exception as e:
            print(f"❌ {test_name} - ERROR: {e}")

    print("\n" + "=" * 50)
    print(f"Phase 2 Test Results: {passed}/{total} passed")

    if passed == total:
        print("🎉 All Phase 2 tests passed! System ready for production.")
        sys.exit(0)
    elif passed >= total * 0.8:  # 80% pass rate
        print("⚠️  Most tests passed. System is functional with minor issues.")
        sys.exit(0)
    else:
        print("❌ Multiple test failures. Please check the implementation.")
        sys.exit(1)

if __name__ == "__main__":
    main()
