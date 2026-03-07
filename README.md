# Proxy RiskScore Checker

## Description
ProxyRiskScoreChecker is a Golang-based command-line utility designed to validate the reliability, anonymity, and security of proxy servers. It verifies the functionality of proxies and subsequently queries the IPQualityScore (IPQS) API to determine their risk or fraud score. Only highly secure and clean proxies (with a risk score of exactly 0) are saved. This tool is invaluable for ensuring your proxy endpoints are not blacklisted, associated with botnets, or otherwise flagged for malicious activity.

## Public API or Exposed Interface

This module is designed as a standalone command-line application. Its primary interfaces are file input/output and environment variables.

### Environment Variables
API_KEY: Your IPQualityScore API key must be set in the environment before execution. 
Example: export API_KEY="your_api_key_here"

### Interactive Prompts
Upon running the program, the user is prompted for:
1. Strictness Level: A value between 0-3 to define how aggressively the IPQS API evaluates the proxy risk. Defaults to 0.
2. Proxy File Name: The text file containing proxies to be evaluated. Defaults to proxies.txt if left blank.

### Input and Output
Input: A newline-separated text file of proxies. Supported protocols: http, https, socks5. Supports proxy authentication.
Output (validproxys.txt): Contains the subset of input proxies that successfully connect and route traffic.
Output (proxies_risk_score_0.txt): Contains the ultimately verified proxies that both connect successfully and yield an IPQS fraud score of 0.

## Relevant Business Logic
The application processes proxies in multiple stages:
1. Parsing and Conversion: Identifies the proxy protocol and authentication details (if any), formatting the string into a standard URL structure. If no protocol is specified, it defaults to http.
2. Live Validation: Connects through the configured proxy to http://ipinfo.io/json to determine the proxy's active outbound IP address. Proxies that timeout or fail to return a valid IP payload are discarded.
3. Risk Scoring: For each successful IP address retrieved, the app calls the IPQualityScore API endpoint using the provided API_KEY and strictness level. 
4. Filtering: Parses the JSON response from IPQS for the fraud_score attribute. Any proxy returning a score greater than 0 is discarded. Only proxies with a fraud_score of 0 are appended to the final output list.
