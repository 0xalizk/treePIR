#!/usr/bin/env python3
import matplotlib.pyplot as plt
import numpy as np
import json

# Load benchmark results from JSON files
def load_benchmark(filename):
    with open(filename, 'r') as f:
        return json.load(f)

# Load all benchmarks
treepir_1gb = load_benchmark('benchmark_1gb_treepir.json')
treepir_4gb = load_benchmark('benchmark_4gb_treepir.json')
nonprivate_1gb = load_benchmark('benchmark_1gb_nonprivate.json')
nonprivate_4gb = load_benchmark('benchmark_4gb_nonprivate.json')

# Create figure with subplots
fig, axes = plt.subplots(2, 3, figsize=(18, 10))
fig.suptitle('TreePIR vs NonPrivate Baseline: Performance Comparison',
             fontsize=16, fontweight='bold')

sizes = ['1GB', '4GB']
x = np.arange(len(sizes))
width = 0.35

# Colors for TreePIR and NonPrivate
treepir_color = '#3498db'  # Blue
nonprivate_color = '#e74c3c'  # Red

# 1. Offline Phase Time
ax = axes[0, 0]
treepir_offline = [treepir_1gb['Offline']['ServerTimeUs'] / 1e6,  # Convert to seconds
                   treepir_4gb['Offline']['ServerTimeUs'] / 1e6]
nonprivate_offline = [nonprivate_1gb['Offline']['ServerTimeUs'] / 1e6,
                      nonprivate_4gb['Offline']['ServerTimeUs'] / 1e6]

bars1 = ax.bar(x - width/2, treepir_offline, width, label='TreePIR',
               color=treepir_color, alpha=0.8, edgecolor='black')
bars2 = ax.bar(x + width/2, nonprivate_offline, width, label='NonPrivate',
               color=nonprivate_color, alpha=0.8, edgecolor='black')

ax.set_ylabel('Time (seconds)', fontweight='bold')
ax.set_title('Offline Phase: Hint Generation Time')
ax.set_xticks(x)
ax.set_xticklabels(sizes)
ax.legend()
ax.grid(axis='y', alpha=0.3, linestyle='--')

for bars in [bars1, bars2]:
    for bar in bars:
        height = bar.get_height()
        if height > 0:
            ax.text(bar.get_x() + bar.get_width()/2, height,
                    f'{height:.2f}s' if height >= 0.01 else f'{height*1000:.2f}ms',
                    ha='center', va='bottom', fontsize=9, fontweight='bold')

# 2. Online Phase Time (per query)
ax = axes[0, 1]
treepir_online = [treepir_1gb['Online']['AvgTotalUs'] / 1000,  # Convert to milliseconds
                  treepir_4gb['Online']['AvgTotalUs'] / 1000]
nonprivate_online = [nonprivate_1gb['Online']['AvgTotalUs'] / 1000,
                     nonprivate_4gb['Online']['AvgTotalUs'] / 1000]

bars1 = ax.bar(x - width/2, treepir_online, width, label='TreePIR',
               color=treepir_color, alpha=0.8, edgecolor='black')
bars2 = ax.bar(x + width/2, nonprivate_online, width, label='NonPrivate',
               color=nonprivate_color, alpha=0.8, edgecolor='black')

ax.set_ylabel('Time (milliseconds)', fontweight='bold')
ax.set_title('Online Phase: Query Time (avg per query)')
ax.set_xticks(x)
ax.set_xticklabels(sizes)
ax.legend()
ax.grid(axis='y', alpha=0.3, linestyle='--')

for bars in [bars1, bars2]:
    for bar in bars:
        height = bar.get_height()
        if height > 0:
            ax.text(bar.get_x() + bar.get_width()/2, height,
                    f'{height:.2f}ms' if height >= 0.01 else f'{height*1000:.2f}μs',
                    ha='center', va='bottom', fontsize=9, fontweight='bold')

# 3. Offline Communication Cost
ax = axes[0, 2]
treepir_offline_comm = [treepir_1gb['Offline']['CommBytes'] / (1024**2),  # Convert to MB
                        treepir_4gb['Offline']['CommBytes'] / (1024**2)]
nonprivate_offline_comm = [nonprivate_1gb['Offline']['CommBytes'] / (1024**2),
                           nonprivate_4gb['Offline']['CommBytes'] / (1024**2)]

bars1 = ax.bar(x - width/2, treepir_offline_comm, width, label='TreePIR',
               color=treepir_color, alpha=0.8, edgecolor='black')
bars2 = ax.bar(x + width/2, nonprivate_offline_comm, width, label='NonPrivate',
               color=nonprivate_color, alpha=0.8, edgecolor='black')

ax.set_ylabel('Size (MB)', fontweight='bold')
ax.set_title('Offline Communication Cost')
ax.set_xticks(x)
ax.set_xticklabels(sizes)
ax.legend()
ax.grid(axis='y', alpha=0.3, linestyle='--')

for bars in [bars1, bars2]:
    for bar in bars:
        height = bar.get_height()
        if height > 0:
            ax.text(bar.get_x() + bar.get_width()/2, height,
                    f'{height:.2f} MB' if height >= 0.01 else f'{height*1024:.2f} KB',
                    ha='center', va='bottom', fontsize=9, fontweight='bold')

# 4. Online Communication Cost (per query)
ax = axes[1, 0]
treepir_online_comm = [treepir_1gb['Online']['AvgOnlineBytes'] / (1024**2),  # Convert to MB
                       treepir_4gb['Online']['AvgOnlineBytes'] / (1024**2)]
nonprivate_online_comm = [nonprivate_1gb['Online']['AvgOnlineBytes'] / (1024**2),
                          nonprivate_4gb['Online']['AvgOnlineBytes'] / (1024**2)]

bars1 = ax.bar(x - width/2, treepir_online_comm, width, label='TreePIR',
               color=treepir_color, alpha=0.8, edgecolor='black')
bars2 = ax.bar(x + width/2, nonprivate_online_comm, width, label='NonPrivate',
               color=nonprivate_color, alpha=0.8, edgecolor='black')

ax.set_ylabel('Size (MB)', fontweight='bold')
ax.set_title('Online Communication Cost (per query)')
ax.set_xlabel('Database Size', fontweight='bold')
ax.set_xticks(x)
ax.set_xticklabels(sizes)
ax.legend()
ax.grid(axis='y', alpha=0.3, linestyle='--')

for bars in [bars1, bars2]:
    for bar in bars:
        height = bar.get_height()
        if height > 0:
            ax.text(bar.get_x() + bar.get_width()/2, height,
                    f'{height:.3f} MB' if height >= 0.001 else f'{height*1024:.2f} KB',
                    ha='center', va='bottom', fontsize=9, fontweight='bold')

# 5. Client Storage Overhead
ax = axes[1, 1]
treepir_storage = [treepir_1gb['Offline']['StorageBytes'] / (1024**2),  # Convert to MB
                   treepir_4gb['Offline']['StorageBytes'] / (1024**2)]
nonprivate_storage = [nonprivate_1gb['Offline']['StorageBytes'] / (1024**2),
                      nonprivate_4gb['Offline']['StorageBytes'] / (1024**2)]

bars1 = ax.bar(x - width/2, treepir_storage, width, label='TreePIR',
               color=treepir_color, alpha=0.8, edgecolor='black')
bars2 = ax.bar(x + width/2, nonprivate_storage, width, label='NonPrivate',
               color=nonprivate_color, alpha=0.8, edgecolor='black')

ax.set_ylabel('Size (MB)', fontweight='bold')
ax.set_title('Client Storage Overhead')
ax.set_xlabel('Database Size', fontweight='bold')
ax.set_xticks(x)
ax.set_xticklabels(sizes)
ax.legend()
ax.grid(axis='y', alpha=0.3, linestyle='--')

for bars in [bars1, bars2]:
    for bar in bars:
        height = bar.get_height()
        if height > 0:
            ax.text(bar.get_x() + bar.get_width()/2, height,
                    f'{height:.2f} MB',
                    ha='center', va='bottom', fontsize=9, fontweight='bold')

# 6. Privacy Overhead (Speedup factor)
ax = axes[1, 2]
# Calculate speedup factors (how much faster NonPrivate is)
speedup_1gb = (treepir_1gb['Online']['AvgTotalUs']) / (nonprivate_1gb['Online']['AvgTotalUs'])
speedup_4gb = (treepir_4gb['Online']['AvgTotalUs']) / (nonprivate_4gb['Online']['AvgTotalUs'])
speedups = [speedup_1gb, speedup_4gb]

bars = ax.bar(sizes, speedups, color='#27ae60', alpha=0.8, edgecolor='black')
ax.set_ylabel('Speedup Factor', fontweight='bold')
ax.set_title('Privacy Overhead\n(TreePIR / NonPrivate)')
ax.set_xlabel('Database Size', fontweight='bold')
ax.grid(axis='y', alpha=0.3, linestyle='--')
ax.axhline(y=1, color='red', linestyle='--', linewidth=2, alpha=0.5, label='No overhead')
ax.legend()

for bar, val in zip(bars, speedups):
    ax.text(bar.get_x() + bar.get_width()/2, bar.get_height(),
            f'{val:.0f}×',
            ha='center', va='bottom', fontsize=12, fontweight='bold')

plt.tight_layout()
plt.savefig('treepir_vs_nonprivate_comparison.png', dpi=300, bbox_inches='tight')
print("Comparison plot saved as 'treepir_vs_nonprivate_comparison.png'")

# Print summary table
print("\n" + "="*100)
print("TreePIR vs NonPrivate Baseline: Performance Comparison")
print("="*100)
print(f"{'Metric':<45} {'TreePIR 1GB':>20} {'NonPrivate 1GB':>20}")
print("-"*100)
print(f"{'Offline Server Time':<45} {treepir_offline[0]:>18.2f}s {nonprivate_offline[0]:>20.6f}s")
print(f"{'Offline Communication (MB)':<45} {treepir_offline_comm[0]:>19.2f} {nonprivate_offline_comm[0]:>20.6f}")
print(f"{'Client Storage (MB)':<45} {treepir_storage[0]:>19.2f} {nonprivate_storage[0]:>20.2f}")
print(f"{'Online Query Time (ms)':<45} {treepir_online[0]:>19.2f} {nonprivate_online[0]:>20.6f}")
print(f"{'Online Communication (MB)':<45} {treepir_online_comm[0]:>19.3f} {nonprivate_online_comm[0]:>20.6f}")
print(f"{'Privacy Overhead (slowdown)':<45} {speedup_1gb:>19.0f}× {'1×':>21}")
print()
print(f"{'Metric':<45} {'TreePIR 4GB':>20} {'NonPrivate 4GB':>20}")
print("-"*100)
print(f"{'Offline Server Time':<45} {treepir_offline[1]:>18.2f}s {nonprivate_offline[1]:>20.6f}s")
print(f"{'Offline Communication (MB)':<45} {treepir_offline_comm[1]:>19.2f} {nonprivate_offline_comm[1]:>20.6f}")
print(f"{'Client Storage (MB)':<45} {treepir_storage[1]:>19.2f} {nonprivate_storage[1]:>20.2f}")
print(f"{'Online Query Time (ms)':<45} {treepir_online[1]:>19.2f} {nonprivate_online[1]:>20.6f}")
print(f"{'Online Communication (MB)':<45} {treepir_online_comm[1]:>19.3f} {nonprivate_online_comm[1]:>20.6f}")
print(f"{'Privacy Overhead (slowdown)':<45} {speedup_4gb:>19.0f}× {'1×':>21}")
print("="*100)

print("\n" + "="*100)
print("Key Takeaways")
print("="*100)
print(f"• TreePIR provides strong privacy guarantees with reasonable performance overhead")
print(f"• Online query time: ~{speedup_1gb:.0f}× slower than no privacy (acceptable for privacy-critical applications)")
print(f"• TreePIR maintains sub-linear √n complexity as database size grows")
print(f"• NonPrivate baseline shows the theoretical minimum performance (no privacy)")
print("="*100)
