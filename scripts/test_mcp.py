#!/usr/bin/env python3
"""
Simple test script for A-MEM MCP Server
Tests the JSON-RPC interface by sending sample requests
"""

import json
import subprocess
import sys
import time

def send_mcp_request(request):
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
        stdout, stderr = process.communicate(input=request_json, timeout=10)
        
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

def test_initialize():
    """Test MCP initialize"""
    print("Testing MCP initialize...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "test-client",
                "version": "1.0.0"
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Initialize response: {json.dumps(response, indent=2)}")
    return response.get("result") is not None

def test_list_tools():
    """Test listing available tools"""
    print("\nTesting list tools...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {}
    }
    
    response = send_mcp_request(request)
    print(f"List tools response: {json.dumps(response, indent=2)}")
    
    if "result" in response and "tools" in response["result"]:
        tools = response["result"]["tools"]
        print(f"Found {len(tools)} tools:")
        for tool in tools:
            print(f"  - {tool['name']}: {tool['description']}")
        return len(tools) > 0
    
    return False

def test_store_memory():
    """Test storing a memory"""
    print("\nTesting store memory...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "store_coding_memory",
            "arguments": {
                "content": "function fibonacci(n) { return n <= 1 ? n : fibonacci(n-1) + fibonacci(n-2); }",
                "project_path": "/test/algorithms",
                "code_type": "javascript",
                "context": "Recursive Fibonacci implementation for testing"
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Store memory response: {json.dumps(response, indent=2)}")
    
    return "result" in response and not response.get("error")

def test_retrieve_memory():
    """Test retrieving memories"""
    print("\nTesting retrieve memory...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 4,
        "method": "tools/call",
        "params": {
            "name": "retrieve_relevant_memories",
            "arguments": {
                "query": "fibonacci algorithm implementation",
                "max_results": 3,
                "min_relevance": 0.5
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Retrieve memory response: {json.dumps(response, indent=2)}")
    
    return "result" in response and not response.get("error")

def test_evolve_network():
    """Test memory network evolution"""
    print("\nTesting evolve network...")
    
    request = {
        "jsonrpc": "2.0",
        "id": 5,
        "method": "tools/call",
        "params": {
            "name": "evolve_memory_network",
            "arguments": {
                "trigger_type": "manual",
                "scope": "recent",
                "max_memories": 10
            }
        }
    }
    
    response = send_mcp_request(request)
    print(f"Evolve network response: {json.dumps(response, indent=2)}")
    
    return "result" in response and not response.get("error")

def main():
    """Run all tests"""
    print("A-MEM MCP Server Test Suite")
    print("=" * 40)
    
    # Check if server binary exists
    try:
        subprocess.run(['./amem-server', '--help'], 
                      capture_output=True, check=False)
    except FileNotFoundError:
        print("Error: amem-server binary not found. Please run 'make build' first.")
        sys.exit(1)
    
    tests = [
        ("Initialize", test_initialize),
        ("List Tools", test_list_tools),
        ("Store Memory", test_store_memory),
        ("Retrieve Memory", test_retrieve_memory),
        ("Evolve Network", test_evolve_network),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        try:
            if test_func():
                print(f"✅ {test_name} - PASSED")
                passed += 1
            else:
                print(f"❌ {test_name} - FAILED")
        except Exception as e:
            print(f"❌ {test_name} - ERROR: {e}")
    
    print("\n" + "=" * 40)
    print(f"Test Results: {passed}/{total} passed")
    
    if passed == total:
        print("🎉 All tests passed!")
        sys.exit(0)
    else:
        print("⚠️  Some tests failed. Check the output above.")
        sys.exit(1)

if __name__ == "__main__":
    main()
