import * as vscode from 'vscode';
import { TestRunner } from './testRunner';
import { TestGenerator } from './testGenerator';
import { CoverageProvider } from './coverageProvider';
import { DebugProvider } from './debugProvider';
import { IntelliSenseProvider } from './intellisenseProvider';
import { TestExplorer } from './testExplorer';
import { DiagnosticsProvider } from './diagnosticsProvider';
import { ConfigurationManager } from './configurationManager';

export function activate(context: vscode.ExtensionContext) {
    console.log('SFX Extension is now active!');
    
    // Initialize providers and managers
    const configManager = new ConfigurationManager();
    const testRunner = new TestRunner(configManager);
    const testGenerator = new TestGenerator(configManager);
    const coverageProvider = new CoverageProvider(context);
    const debugProvider = new DebugProvider();
    const intellisenseProvider = new IntelliSenseProvider();
    const testExplorer = new TestExplorer(context, testRunner);
    const diagnosticsProvider = new DiagnosticsProvider();
    
    // Register commands
    registerCommands(context, testRunner, testGenerator, coverageProvider, debugProvider, configManager);
    
    // Register providers
    registerProviders(context, intellisenseProvider, coverageProvider, diagnosticsProvider);
    
    // Register views
    registerViews(context, testExplorer);
    
    // Setup workspace monitoring
    setupWorkspaceMonitoring(context, testExplorer, diagnosticsProvider);
    
    // Show welcome message for first-time users
    showWelcomeMessage(context);
}

function registerCommands(
    context: vscode.ExtensionContext,
    testRunner: TestRunner,
    testGenerator: TestGenerator,
    coverageProvider: CoverageProvider,
    debugProvider: DebugProvider,
    configManager: ConfigurationManager
) {
    // Test execution commands
    context.subscriptions.push(
        vscode.commands.registerCommand('sfx.test.run', () => testRunner.runAllTests()),
        vscode.commands.registerCommand('sfx.test.runFile', () => testRunner.runTestsInFile()),
        vscode.commands.registerCommand('sfx.test.runSingle', () => testRunner.runSingleTest()),
        vscode.commands.registerCommand('sfx.test.debug', () => debugProvider.debugTest())
    );
    
    // Test generation commands
    context.subscriptions.push(
        vscode.commands.registerCommand('sfx.generate.test', () => testGenerator.generateTest()),
        vscode.commands.registerCommand('sfx.generate.scaffold', (uri) => testGenerator.scaffoldTestSuite(uri))
    );
    
    // Analysis commands
    context.subscriptions.push(
        vscode.commands.registerCommand('sfx.coverage.show', () => coverageProvider.showCoverage()),
        vscode.commands.registerCommand('sfx.benchmark.run', () => testRunner.runBenchmarks()),
        vscode.commands.registerCommand('sfx.logs.analyze', () => debugProvider.analyzeLogs()),
        vscode.commands.registerCommand('sfx.trace.analyze', () => debugProvider.analyzeTraces()),
        vscode.commands.registerCommand('sfx.config.validate', () => configManager.validateConfiguration()),
        vscode.commands.registerCommand('sfx.doctor', () => runSfxDoctor())
    );
}

function registerProviders(
    context: vscode.ExtensionContext,
    intellisenseProvider: IntelliSenseProvider,
    coverageProvider: CoverageProvider,
    diagnosticsProvider: DiagnosticsProvider
) {
    // IntelliSense provider for Go files
    context.subscriptions.push(
        vscode.languages.registerCompletionItemProvider(
            { scheme: 'file', language: 'go' },
            intellisenseProvider,
            '.',
            '('
        )
    );
    
    // Hover provider for documentation
    context.subscriptions.push(
        vscode.languages.registerHoverProvider(
            { scheme: 'file', language: 'go' },
            intellisenseProvider
        )
    );
    
    // Definition provider for navigation
    context.subscriptions.push(
        vscode.languages.registerDefinitionProvider(
            { scheme: 'file', language: 'go' },
            intellisenseProvider
        )
    );
    
    // Code action provider for quick fixes
    context.subscriptions.push(
        vscode.languages.registerCodeActionsProvider(
            { scheme: 'file', language: 'go' },
            intellisenseProvider,
            {
                providedCodeActionKinds: [
                    vscode.CodeActionKind.QuickFix,
                    vscode.CodeActionKind.Refactor
                ]
            }
        )
    );
    
    // Diagnostics provider for linting
    context.subscriptions.push(diagnosticsProvider);
}

function registerViews(context: vscode.ExtensionContext, testExplorer: TestExplorer) {
    // Test explorer view
    const testTreeProvider = testExplorer.getTreeDataProvider();
    vscode.window.registerTreeDataProvider('sfxTestExplorer', testTreeProvider);
    
    // Coverage view
    const coverageTreeProvider = new CoverageTreeProvider();
    vscode.window.registerTreeDataProvider('sfxCoverage', coverageTreeProvider);
    
    // Register view commands
    context.subscriptions.push(
        vscode.commands.registerCommand('sfxTestExplorer.refresh', () => testTreeProvider.refresh()),
        vscode.commands.registerCommand('sfxTestExplorer.runTest', (item) => testExplorer.runTest(item)),
        vscode.commands.registerCommand('sfxCoverage.refresh', () => coverageTreeProvider.refresh())
    );
}

function setupWorkspaceMonitoring(
    context: vscode.ExtensionContext,
    testExplorer: TestExplorer,
    diagnosticsProvider: DiagnosticsProvider
) {
    // Watch for Go file changes
    const fileWatcher = vscode.workspace.createFileSystemWatcher('**/*.go');
    
    fileWatcher.onDidChange((uri) => {
        diagnosticsProvider.validateFile(uri);
        testExplorer.refreshTests();
    });
    
    fileWatcher.onDidCreate((uri) => {
        if (uri.fsPath.endsWith('_test.go')) {
            testExplorer.refreshTests();
        }
        // Auto-generate test template if enabled
        const config = vscode.workspace.getConfiguration('sfx');
        if (config.get('autoGenerate.enabled')) {
            suggestTestGeneration(uri);
        }
    });
    
    fileWatcher.onDidDelete((uri) => {
        testExplorer.refreshTests();
    });
    
    context.subscriptions.push(fileWatcher);
    
    // Watch for configuration changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration((e) => {
            if (e.affectsConfiguration('sfx')) {
                vscode.window.showInformationMessage('SFX configuration changed. Some features may require restart.');
            }
        })
    );
}

async function showWelcomeMessage(context: vscode.ExtensionContext) {
    const showWelcome = context.globalState.get('sfx.showWelcome', true);
    
    if (showWelcome) {
        const action = await vscode.window.showInformationMessage(
            'Welcome to SFX! Would you like to see the getting started guide?',
            'Show Guide',
            'Not Now',
            "Don't Show Again"
        );
        
        switch (action) {
            case 'Show Guide':
                vscode.commands.executeCommand('vscode.open', vscode.Uri.parse('https://sfx.dev/getting-started'));
                break;
            case "Don't Show Again":
                context.globalState.update('sfx.showWelcome', false);
                break;
        }
    }
}

async function suggestTestGeneration(uri: vscode.Uri) {
    if (uri.fsPath.endsWith('_test.go')) {
        return; // Already a test file
    }
    
    const action = await vscode.window.showInformationMessage(
        `Generate tests for ${uri.fsPath}?`,
        'Generate',
        'Not Now'
    );
    
    if (action === 'Generate') {
        vscode.commands.executeCommand('sfx.generate.test', uri);
    }
}

async function runSfxDoctor() {
    const terminal = vscode.window.createTerminal('SFX Doctor');
    terminal.sendText('sfx doctor --verbose');
    terminal.show();
}

class CoverageTreeProvider implements vscode.TreeDataProvider<CoverageItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<CoverageItem | undefined | null | void> = new vscode.EventEmitter<CoverageItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<CoverageItem | undefined | null | void> = this._onDidChangeTreeData.event;
    
    refresh(): void {
        this._onDidChangeTreeData.fire();
    }
    
    getTreeItem(element: CoverageItem): vscode.TreeItem {
        return element;
    }
    
    getChildren(element?: CoverageItem): Thenable<CoverageItem[]> {
        if (!element) {
            // Root items - show package coverage
            return Promise.resolve([
                new CoverageItem('lambda', '92.4%', vscode.TreeItemCollapsibleState.Collapsed),
                new CoverageItem('dynamodb', '89.7%', vscode.TreeItemCollapsibleState.Collapsed),
                new CoverageItem('eventbridge', '95.1%', vscode.TreeItemCollapsibleState.Collapsed),
                new CoverageItem('stepfunctions', '88.3%', vscode.TreeItemCollapsibleState.Collapsed)
            ]);
        } else {
            // Child items - show file coverage
            return Promise.resolve([
                new CoverageItem(`${element.label}.go`, '90.0%', vscode.TreeItemCollapsibleState.None),
                new CoverageItem(`${element.label}_test.go`, '100%', vscode.TreeItemCollapsibleState.None)
            ]);
        }
    }
}

class CoverageItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly coverage: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(label, collapsibleState);
        this.tooltip = `${this.label} - Coverage: ${this.coverage}`;
        this.description = this.coverage;
        
        // Set icon based on coverage percentage
        const percent = parseFloat(coverage.replace('%', ''));
        if (percent >= 90) {
            this.iconPath = new vscode.ThemeIcon('pass', new vscode.ThemeColor('testing.iconPassed'));
        } else if (percent >= 70) {
            this.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('testing.iconQueued'));
        } else {
            this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('testing.iconFailed'));
        }
    }
}

export function deactivate() {
    console.log('SFX Extension is now deactivated');
}