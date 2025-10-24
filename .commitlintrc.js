// Commitlint 配置
// 支持中英文提交信息

module.exports = {
  // 继承默认配置
  extends: ['@commitlint/config-conventional'],
  
  // 自定义规则
  rules: {
    // 类型枚举（支持中英文）
    'type-enum': [
      2,
      'always',
      [
        // 英文类型（标准）
        'feat',      // 新功能
        'fix',       // Bug修复
        'docs',      // 文档更新
        'style',     // 代码格式（不影响逻辑）
        'refactor',  // 重构
        'perf',      // 性能优化
        'test',      // 测试相关
        'build',     // 构建系统或依赖变更
        'ci',        // CI配置变更
        'chore',     // 其他变更
        'revert',    // 回滚
        // 中文类型（可选，推荐使用英文）
        '新功能',
        '修复',
        '文档',
        '格式',
        '重构',
        '性能',
        '测试',
        '构建',
        '集成',
        '杂项',
        '回滚'
      ]
    ],
    
    // 类型大小写
    'type-case': [2, 'always', 'lower-case'],
    
    // 类型不能为空
    'type-empty': [2, 'never'],
    
    // 主题不能为空
    'subject-empty': [2, 'never'],
    
    // 主题不能以句号结尾
    'subject-full-stop': [2, 'never', '.'],
    
    // 主题大小写（禁用，允许任意大小写）
    'subject-case': [0],
    
    // Header 最大长度
    'header-max-length': [2, 'always', 100],
    
    // Body 前导空行
    'body-leading-blank': [1, 'always'],
    
    // Footer 前导空行
    'footer-leading-blank': [1, 'always']
  },
  
  // 提示信息
  prompt: {
    messages: {
      type: '选择你要提交的类型 :',
      scope: '选择一个 scope（可选）:',
      customScope: '请输入自定义的 scope :',
      subject: '填写简短精炼的变更描述 :\n',
      body: '填写更加详细的变更描述（可选）。使用 "|" 换行 :\n',
      breaking: '列举非兼容性重大的变更（可选）。使用 "|" 换行 :\n',
      footerPrefixesSelect: '选择关联issue前缀（可选）:',
      customFooterPrefix: '输入自定义issue前缀 :',
      footer: '列举关联issue（可选）例如: #31, #I3244 :\n',
      confirmCommit: '是否提交或修改commit ?'
    },
    types: [
      { value: 'feat', name: 'feat:     ✨  新功能', emoji: ':sparkles:' },
      { value: 'fix', name: 'fix:      🐛  Bug修复', emoji: ':bug:' },
      { value: 'docs', name: 'docs:     📝  文档更新', emoji: ':memo:' },
      { value: 'style', name: 'style:    💄  代码格式', emoji: ':lipstick:' },
      { value: 'refactor', name: 'refactor: ♻️   代码重构', emoji: ':recycle:' },
      { value: 'perf', name: 'perf:     ⚡️  性能优化', emoji: ':zap:' },
      { value: 'test', name: 'test:     ✅  测试相关', emoji: ':white_check_mark:' },
      { value: 'build', name: 'build:    📦️  构建系统', emoji: ':package:' },
      { value: 'ci', name: 'ci:       🎡  CI配置', emoji: ':ferris_wheel:' },
      { value: 'chore', name: 'chore:    🔨  其他变更', emoji: ':hammer:' },
      { value: 'revert', name: 'revert:   ⏪️  回滚', emoji: ':rewind:' }
    ],
    useEmoji: false,
    emojiAlign: 'center',
    allowCustomScopes: true,
    allowEmptyScopes: true,
    customScopesAlign: 'bottom',
    customScopesAlias: 'custom',
    emptyScopesAlias: 'empty',
    upperCaseSubject: false,
    markBreakingChangeMode: false,
    allowBreakingChanges: ['feat', 'fix'],
    breaklineNumber: 100,
    breaklineChar: '|',
    skipQuestions: [],
    issuePrefixes: [
      { value: 'closed', name: 'closed:   已关闭' },
      { value: 'fixes', name: 'fixes:    修复' }
    ],
    customIssuePrefixAlign: 'top',
    emptyIssuePrefixAlias: 'skip',
    customIssuePrefixAlias: 'custom',
    allowCustomIssuePrefix: true,
    allowEmptyIssuePrefix: true,
    confirmColorize: true,
    maxHeaderLength: Infinity,
    maxSubjectLength: Infinity,
    minSubjectLength: 0,
    scopeOverrides: undefined,
    defaultBody: '',
    defaultIssues: '',
    defaultScope: '',
    defaultSubject: ''
  }
};

