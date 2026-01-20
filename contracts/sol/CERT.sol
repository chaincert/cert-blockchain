// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title IERC20
 * @notice Standard ERC-20 Interface
 */
interface IERC20 {
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function approve(address spender, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}

/**
 * @title IERC20Metadata
 * @notice ERC-20 Metadata Extension
 */
interface IERC20Metadata is IERC20 {
    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
}

/**
 * @title CERT
 * @notice CERT Token ERC-20 Implementation for EVM compatibility
 * @dev Per CERT Whitepaper Section 5
 * 
 * Token Parameters:
 * - Symbol: CERT
 * - Total Supply: 1,000,000,000 (1 Billion) - Fixed, Non-inflationary
 * - Decimals: 6 (1 CERT = 1,000,000 ucert)
 * 
 * This is a wrapped representation of the native CERT token
 * for EVM contract interactions. The native token uses the Cosmos
 * SDK bank module with denomination 'ucert'.
 */
contract CERT is IERC20, IERC20Metadata {
    // Token metadata per Whitepaper Section 5.1
    string private constant _name = "CERT";
    string private constant _symbol = "CERT";
    uint8 private constant _decimals = 6;
    
    // Total supply: 1 Billion CERT (Whitepaper Section 5.1)
    // 1,000,000,000 * 10^6 = 1,000,000,000,000,000 (1 quadrillion ucert)
    uint256 private constant _totalSupply = 1_000_000_000 * 10**6;

    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    // Bridge address for native <-> ERC20 conversions
    address public immutable bridge;

    event BridgeMint(address indexed to, uint256 amount);
    event BridgeBurn(address indexed from, uint256 amount);

    modifier onlyBridge() {
        require(msg.sender == bridge, "Only bridge can call");
        _;
    }

    constructor(address _bridge) {
        require(_bridge != address(0), "Invalid bridge address");
        bridge = _bridge;
    }

    function name() external pure override returns (string memory) {
        return _name;
    }

    function symbol() external pure override returns (string memory) {
        return _symbol;
    }

    function decimals() external pure override returns (uint8) {
        return _decimals;
    }

    function totalSupply() external pure override returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) external view override returns (uint256) {
        return _balances[account];
    }

    function transfer(address to, uint256 amount) external override returns (bool) {
        _transfer(msg.sender, to, amount);
        return true;
    }

    function allowance(address owner, address spender) external view override returns (uint256) {
        return _allowances[owner][spender];
    }

    function approve(address spender, uint256 amount) external override returns (bool) {
        _approve(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external override returns (bool) {
        _spendAllowance(from, msg.sender, amount);
        _transfer(from, to, amount);
        return true;
    }

    /**
     * @notice Mint wrapped CERT tokens (only bridge)
     * @dev Called when native CERT is deposited to the bridge
     */
    function mint(address to, uint256 amount) external onlyBridge {
        require(to != address(0), "Mint to zero address");
        _balances[to] += amount;
        emit Transfer(address(0), to, amount);
        emit BridgeMint(to, amount);
    }

    /**
     * @notice Burn wrapped CERT tokens (only bridge)
     * @dev Called when withdrawing to native CERT
     */
    function burn(address from, uint256 amount) external onlyBridge {
        require(_balances[from] >= amount, "Insufficient balance");
        _balances[from] -= amount;
        emit Transfer(from, address(0), amount);
        emit BridgeBurn(from, amount);
    }

    function _transfer(address from, address to, uint256 amount) internal {
        require(from != address(0), "Transfer from zero address");
        require(to != address(0), "Transfer to zero address");
        require(_balances[from] >= amount, "Insufficient balance");

        unchecked {
            _balances[from] -= amount;
            _balances[to] += amount;
        }

        emit Transfer(from, to, amount);
    }

    function _approve(address owner, address spender, uint256 amount) internal {
        require(owner != address(0), "Approve from zero address");
        require(spender != address(0), "Approve to zero address");

        _allowances[owner][spender] = amount;
        emit Approval(owner, spender, amount);
    }

    function _spendAllowance(address owner, address spender, uint256 amount) internal {
        uint256 currentAllowance = _allowances[owner][spender];
        if (currentAllowance != type(uint256).max) {
            require(currentAllowance >= amount, "Insufficient allowance");
            unchecked {
                _allowances[owner][spender] = currentAllowance - amount;
            }
        }
    }
}

