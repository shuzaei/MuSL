document.addEventListener('DOMContentLoaded', function () {
    // スライダー値の表示更新
    document.getElementById('majorProb').addEventListener('input', function () {
        document.getElementById('majorProbValue').textContent = this.value;
    });

    // 高度な設定のスライダー値の表示更新
    document.getElementById('innovationRate').addEventListener('input', function () {
        document.getElementById('innovationRateValue').textContent = this.value;
    });

    document.getElementById('noveltyPreference').addEventListener('input', function () {
        document.getElementById('noveltyPreferenceValue').textContent = this.value;
    });

    document.getElementById('reproductionProbability').addEventListener('input', function () {
        document.getElementById('reproductionProbabilityValue').textContent = this.value;
    });

    document.getElementById('creationProbability').addEventListener('input', function () {
        document.getElementById('creationProbabilityValue').textContent = this.value;
    });

    document.getElementById('listeningProbability').addEventListener('input', function () {
        document.getElementById('listeningProbabilityValue').textContent = this.value;
    });

    // タブ切り替え
    // ... existing code ...

    // シミュレーション実行
    function runSimulation() {
        const majorProb = parseFloat(document.getElementById('majorProb').value);
        const initialAgents = parseInt(document.getElementById('initialAgents').value);
        const iterations = parseInt(document.getElementById('iterations').value);

        // 高度な設定のパラメータを取得
        const innovationRate = parseFloat(document.getElementById('innovationRate').value);
        const noveltyPreference = parseFloat(document.getElementById('noveltyPreference').value);
        const reproductionProbability = parseFloat(document.getElementById('reproductionProbability').value);
        const creationProbability = parseFloat(document.getElementById('creationProbability').value);
        const listeningProbability = parseFloat(document.getElementById('listeningProbability').value);

        const requestData = {
            majorProb: majorProb,
            initialAgents: initialAgents,
            iterations: iterations,
            advancedSettings: {
                innovationRate: innovationRate,
                noveltyPreference: noveltyPreference,
                reproductionProbability: reproductionProbability,
                creationProbability: creationProbability,
                listeningProbability: listeningProbability
            }
        };

        updateProgress('simulation', 0, 'シミュレーション開始中...');
        // ... existing code ...
    }
}); 